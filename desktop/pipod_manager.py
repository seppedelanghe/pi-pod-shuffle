import argparse
import os
import traceback
import torch
import subprocess
import numpy as np

from viz import cmd_visualize
from helpers import load_json, save_json
from extraction import extract_embeddings
from statics import RAW_DB_FILE, LIBRARY_FILE, DEFAULT_MUSIC_DIR, SUPPORTED_EXTS

from pytorch.models import Cnn6


# Try imports and handle missing libraries gracefully
try:
    import librosa
    LIBS_AVAILABLE = True
except ImportError as e:
    print(f"Warning: Audio/ML libraries not found ({e}). 'process' command will fail.")
    print("Install them with: pip install librosa numpy")
    LIBS_AVAILABLE = False


# --- HELPERS ---


# --- COMMANDS ---

def cmd_scan(args):
    """Scans disk vs raw_data.json to find new files."""
    print(f"--- Scanning {args.dir} ---")
    
    raw_data = load_json(os.path.join(args.dir, RAW_DB_FILE))
    
    # 1. Walk Disk
    files_on_disk = set()
    for root, _, files in os.walk(args.dir):
        for file in files:
            if file.lower().endswith(SUPPORTED_EXTS):
                # Rel path ensures portability
                full_path = os.path.join(root, file)
                rel_path = os.path.relpath(full_path, args.dir)
                files_on_disk.add(rel_path)

    # 2. Compare with DB
    files_in_db = set(raw_data.keys())
    
    new_files = files_on_disk - files_in_db
    deleted_files = files_in_db - files_on_disk

    # 3. Handle Deletions
    if deleted_files:
        print(f"Found {len(deleted_files)} deleted files. Removing from DB...")
        for f in deleted_files:
            del raw_data[f]
        save_json(raw_data, os.path.join(args.dir, RAW_DB_FILE))
    
    print("Scan complete.")
    print(f"Total files: {len(files_on_disk)}")
    print(f"New files to process: {len(new_files)}")
    
    if new_files:
        print("\nRun 'python pipod_manager.py process' to analyze them.")


def cmd_process(args):
    """Analyzes new files and re-runs PCA."""
    if not LIBS_AVAILABLE:
        print("Error: Missing required libraries (librosa, sklearn).")
        return

    raw_db_path = os.path.join(args.dir, RAW_DB_FILE)
    lib_db_path = os.path.join(args.dir, LIBRARY_FILE)
    
    raw_data = load_json(raw_db_path)
    
    # Find what needs processing
    files_on_disk = set()
    for root, _, files in os.walk(args.dir):
        for file in files:
            if file.lower().endswith(SUPPORTED_EXTS):
                rel_path = os.path.relpath(os.path.join(root, file), args.dir)
                files_on_disk.add(rel_path)
    
    to_process = files_on_disk - set(raw_data.keys())

    # load ML Model
    device = 'cuda' if torch.cuda.is_available() else 'mps'

    # Initialize the architecture
    model = Cnn6(sample_rate=32000, window_size=1024, 
                 hop_size=320, mel_bins=64, fmin=50, 
                 fmax=14000, classes_num=527)

    # Load your weights
    checkpoint = torch.load("Cnn6_mAP=0.343.pth", map_location=device)
    model.load_state_dict(checkpoint['model'])
    model.to(device)
    model.eval()
    
    if not to_process and os.path.exists(lib_db_path):
        print("No new files to analyze.")
    else:
        print(f"--- Processing {len(to_process)} New Files ---")
        count = 0
        for i, rel_path in enumerate(to_process):
            full_path = os.path.join(args.dir, rel_path)
            print(f"[{i+1}/{len(to_process)}] Analyzing: {rel_path}...")
            
            try:
                vec = extract_embeddings(full_path, model)
            except Exception as e:
                print(f"\n[!] SKIPPING CORRUPT FILE: {full_path}")
                print(f"Reason: {e}")
                print(traceback.format_exc())
                continue

            raw_data[rel_path] = vec[0]
            count += 1

        print("Finished analysis. Saving raw data...")
        save_json(raw_data, raw_db_path)

    # Create final Library JSON
    foldername = os.path.basename(os.path.normpath(args.dir))
    library_data = {"dir": f'/home/pipod/{foldername}', "files": {}}
    for i, path in enumerate(raw_data.keys()):
        # Rounding saves space and is fine for similarity checks
        library_data['files'][path] = np.round(raw_data[path], 4).tolist()

    save_json(library_data, lib_db_path)
    print("Ready to sync.")


def cmd_sync(args):
    """Syncs files and library.json to Pi via rsync."""
    if not args.user or not args.ip or not args.dest:
        print("Error: Sync requires --user, --ip, and --dest arguments.")
        return

    remote = f"{args.user}@{args.ip}:{args.dest}"
    print(f"--- Syncing to {remote} ---")

    # Ensure local path ends in / for rsync to copy CONTENTS, not folder
    src = args.dir if args.dir.endswith('/') else args.dir + '/'

    # 1. Sync Music (exclude raw data file, include everything else)
    # --delete removes songs on Pi that were deleted locally
    cmd = [
        "rsync", "-av", "--delete",
        "--exclude", RAW_DB_FILE,  # Don't send the big raw file
        src, remote
    ]
    
    print("Running rsync...")
    try:
        subprocess.run(cmd, check=True)
        print("Sync Successful!")
    except subprocess.CalledProcessError as e:
        print(f"Sync Failed: {e}")
    except FileNotFoundError:
        print("Error: 'rsync' command not found on this system.")

# --- MAIN ---

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Pi Pod Music Manager")
    parser.add_argument("--dir", default=DEFAULT_MUSIC_DIR, help="Path to local music folder")
    
    subparsers = parser.add_subparsers(dest="command", required=True)

    # Scan
    p_scan = subparsers.add_parser("scan", help="Check for new files")

    # Process
    p_process = subparsers.add_parser("process", help="Analyze audio and update library")
    
    # Process
    p_viz = subparsers.add_parser("viz", help="Visualize the analyzed audio")
    p_viz.add_argument(
        "--anchor",
        help="Anchor song (partial filename match, case-insensitive)"
    )

    # Sync
    p_sync = subparsers.add_parser("sync", help="Sync to Pi")
    p_sync.add_argument("--user", help="Pi SSH Username (e.g. pi)")
    p_sync.add_argument("--ip", help="Pi IP Address")
    p_sync.add_argument("--dest", help="Remote destination path (e.g. /home/pi/music)")

    args = parser.parse_args()

    if args.command == "scan":
        cmd_scan(args)
    elif args.command == "process":
        cmd_process(args)
    elif args.command == "viz":
        cmd_visualize(args)
    elif args.command == "sync":
        cmd_sync(args)
