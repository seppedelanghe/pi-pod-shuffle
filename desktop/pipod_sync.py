import customtkinter as ctk
from tkinter import filedialog
import os
import json
import threading
import subprocess
import numpy as np

# --- ESSENTIA IMPORTS ---
# Wrap in try/except to allow GUI to open even if installation is broken
try:
    import essentia.standard as es
    ESSENTIA_AVAILABLE = True
except ImportError as e:
    print(f"Warning: Essentia not found. Analysis will fail. Error: {e}")
    ESSENTIA_AVAILABLE = False

# --- CONFIGURATION ---
MODEL_PB = os.path.join("models", "discogs-effnet-bs64-1.pb")
MODEL_JSON = os.path.join("models", "discogs-effnet-bs64-1.json")
SUPPORTED_EXTS = ('.mp3', '.flac')
LIBRARY_FILENAME = "library.json"

ctk.set_appearance_mode("Dark")
ctk.set_default_color_theme("blue")


class PiPodSyncApp(ctk.CTk):
    def __init__(self):
        super().__init__()

        self.title("Pi Pod Sync Manager")
        self.geometry("700x600")

        # Data states
        self.local_music_path = ""
        self.db_data = {}
        self.files_to_process = set()
        self.is_processing = False

        self._setup_gui()

    def _setup_gui(self):
        # === Top Frame: Configuration ===
        self.config_frame = ctk.CTkFrame(self)
        self.config_frame.pack(pady=10, padx=10, fill="x")

        ctk.CTkLabel(self.config_frame, text="Configuration", font=("Arial", 16, "bold")).pack(pady=(10,5))

        # Local Path Checkbox
        path_row = ctk.CTkFrame(self.config_frame, fg_color="transparent")
        path_row.pack(fill="x", padx=10, pady=5)
        self.path_label = ctk.CTkLabel(path_row, text="Local Music Folder: Not selected", anchor="w")
        self.path_label.pack(side="left", fill="x", expand=True)
        btn_browse = ctk.CTkButton(path_row, text="Browse", width=100, command=self.select_folder)
        btn_browse.pack(side="right")

        # Pi Config
        pi_row = ctk.CTkFrame(self.config_frame, fg_color="transparent")
        pi_row.pack(fill="x", padx=10, pady=5)
        self.pi_user_entry = ctk.CTkEntry(pi_row, placeholder_text="pi user (e.g., pi)", width=150)
        self.pi_user_entry.pack(side="left", padx=(0, 10))
        self.pi_ip_entry = ctk.CTkEntry(pi_row, placeholder_text="Pi IP Address (e.g., 192.168.1.50)", width=200)
        self.pi_ip_entry.pack(side="left", padx=(0, 10))
        self.pi_dest_entry = ctk.CTkEntry(pi_row, placeholder_text="Dest Path (e.g., /home/pi/music/)", width=200)
        self.pi_dest_entry.pack(side="left")

        # === Middle Frame: Actions ===
        self.action_frame = ctk.CTkFrame(self)
        self.action_frame.pack(pady=10, padx=10, fill="x")

        ctk.CTkLabel(self.action_frame, text="Actions", font=("Arial", 16, "bold")).pack(pady=(10, 5))
        
        btn_row = ctk.CTkFrame(self.action_frame, fg_color="transparent")
        btn_row.pack(pady=10)

        self.btn_scan = ctk.CTkButton(btn_row, text="1. Scan Folder & Cleanup DB", command=self.start_scan, width=200, height=40, fg_color="green")
        self.btn_scan.pack(side="left", padx=10)
        self.btn_scan.configure(state="disabled")

        self.btn_process = ctk.CTkButton(btn_row, text="2. Process New Audio (AI)", command=self.start_processing_thread, width=200, height=40)
        self.btn_process.pack(side="left", padx=10)
        self.btn_process.configure(state="disabled")

        self.btn_sync = ctk.CTkButton(btn_row, text="3. Sync to Pi (rsync)", command=self.start_sync_thread, width=200, height=40, fg_color="#D35400")
        self.btn_sync.pack(side="left", padx=10)
        self.btn_sync.configure(state="disabled")
        
        self.progress_bar = ctk.CTkProgressBar(self.action_frame)
        self.progress_bar.pack(fill="x", padx=20, pady=(0, 20))
        self.progress_bar.set(0)

        # === Bottom Frame: Logs ===
        self.log_box = ctk.CTkTextbox(self)
        self.log_box.pack(pady=10, padx=10, fill="both", expand=True)
        self.log("Welcome to Pi Pod Sync. Please select your local music folder.")

        if not ESSENTIA_AVAILABLE:
             self.log("ERROR: Essentia library not loaded. Processing will not work.", "red")
        
        if not os.path.exists(MODEL_PB):
             self.log(f"ERROR: Model file missing at {MODEL_PB}. Please download it.", "red")

    # --- HELPER FUNCTIONS ---
    def log(self, message, color=None):
        self.log_box.configure(state="normal")
        tag = None
        if color:
            tag = color
            self.log_box.tag_config(color, foreground=color)
        
        self.log_box.insert("end", message + "\n", tag)
        self.log_box.see("end")
        self.log_box.configure(state="disabled")

    def select_folder(self):
        path = filedialog.askdirectory()
        if path:
            self.local_music_path = path
            self.path_label.configure(text=f"Local Music Folder: {self.local_music_path}")
            self.btn_scan.configure(state="normal")
            self.log(f"Selected folder: {path}")
            # Reset states
            self.files_to_process = set()
            self.btn_process.configure(state="disabled")
            self.btn_sync.configure(state="disabled")

    # --- STEP 1: SCANNING LOGIC ---
    def start_scan(self):
        self.log("Scanning folder and checking database integrity...")
        db_path = os.path.join(self.local_music_path, LIBRARY_FILENAME)

        # 1. Load existing DB
        if os.path.exists(db_path):
            try:
                with open(db_path, 'r') as f:
                    self.db_data = json.load(f)
                self.log(f"Loaded existing database with {len(self.db_data)} songs.")
            except Exception as e:
                self.log(f"Error loading DB file, starting fresh: {e}", "red")
                self.db_data = {}
        else:
            self.log("No existing database found. Starting fresh.")
            self.db_data = {}

        # 2. Scan Disk for MP3 and FLAC
        files_on_disk = set()
        for root, dirs, files in os.walk(self.local_music_path):
            for file in files:
                if file.lower().endswith(SUPPORTED_EXTS):
                    # Store relative path to keep it portable
                    rel_path = os.path.relpath(os.path.join(root, file), self.local_music_path)
                    files_on_disk.add(rel_path)

        files_in_db = set(self.db_data.keys())

        # 3. Calculate Deltas
        self.files_to_process = files_on_disk - files_in_db
        deleted_files = files_in_db - files_on_disk

        # 4. Handle Deletions (Cleanup)
        if deleted_files:
            self.log(f"Cleanup: Removing {len(deleted_files)} obsolete entries from DB.", "orange")
            for f in deleted_files:
                del self.db_data[f]
            # Save cleaned DB immediately
            self.save_db()
        else:
            self.log("Database clean (no deleted files found).")

        # 5. Report Results
        if self.files_to_process:
            self.log(f"Scan Complete: Found {len(self.files_to_process)} NEW files needing analysis.", "green")
            self.btn_process.configure(state="normal", text=f"2. Process {len(self.files_to_process)} New Files (AI)")
        else:
            self.log("Scan Complete: No new files found. Database up to date.")
            self.btn_process.configure(state="disabled", text="2. Process New Audio (AI)")
            self.btn_sync.configure(state="normal")

    def save_db(self):
        db_path = os.path.join(self.local_music_path, LIBRARY_FILENAME)
        with open(db_path, 'w') as f:
             # Use indent for readability, though it makes file larger
            json.dump(self.db_data, f, indent=2)

    # --- STEP 2: AI PROCESSING LOGIC ---

    def analyze_audio_file(self, full_path):
        """Run Essentia inference on a single file."""
        if not ESSENTIA_AVAILABLE: return None

        # Initialize algorithms
        # MonoLoader handles resampling to 16kHz automatically, which models usually require
        loader = es.MonoLoader(filename=full_path, sampleRate=16000, resampleQuality=1)
        
        # The model expects specific input size. We use TensorflowPredict directly.
        # We will slice the audio into patches.
        tf_predict = es.TensorflowPredict(graphFilename=MODEL_PB, 
                                          inputs=["serving_default_model_input"], 
                                          outputs=["StatefulPartitionedCall"],
                                          squeeze=True) # Ensure output is squeezed

        audio = loader()
        
        # Ensure audio is long enough for at least one patch (approx 3-5 seconds depending on model)
        # Discogs-Effnet usually wants ~3 seconds chunks. 
        # If too short, pad with zeros.
        target_patch_size = 48000 # approx 3 seconds at 16khz
        if len(audio) < target_patch_size:
             audio = np.pad(audio, (0, target_patch_size - len(audio)))

        # Run prediction. This model often returns embeddings for patches.
        # We need to aggregate them to get one vector for the song.
        try:
            # This returns a matrix [N_patches x VectorDim]
            raw_embeddings = tf_predict(audio)
            
            # Simple aggregation: Average all patches to get one representative vector.
            # Axis 0 means average down the rows.
            averaged_vector = np.mean(raw_embeddings, axis=0)
            
            # Convert numpy floats to regular floats for JSON serialization
            return [float(x) for x in averaged_vector]

        except Exception as e:
            print(f"Essentia Error on file {full_path}: {e}")
            return None


    def start_processing_thread(self):
        if self.is_processing: return
        self.is_processing = True
        self.btn_process.configure(state="disabled")
        self.btn_scan.configure(state="disabled")
        self.btn_sync.configure(state="disabled")
        
        threading.Thread(target=self.process_loop, daemon=True).start()

    def process_loop(self):
        total = len(self.files_to_process)
        current = 0
        failed = 0
        self.log(f"Starting AI Analysis on {total} files. This will take time...")

        # Iterate over a copy so we can modify set if needed
        files_list = list(self.files_to_process)

        for rel_path in files_list:
            current += 1
            full_path = os.path.join(self.local_music_path, rel_path)
            self.log(f"[{current}/{total}] Analyzing: {rel_path}...")
            self.progress_bar.set(current / total)

            try:
                vector = self.analyze_audio_file(full_path)
                if vector and len(vector) > 0:
                    self.db_data[rel_path] = vector
                else:
                    self.log(f"Failed to get vector for {rel_path}", "orange")
                    failed += 1
            except Exception as e:
                self.log(f"CRITICAL ERROR analyzing {rel_path}: {e}", "red")
                failed += 1

        # Finish up
        self.save_db()
        self.log(f"Processing complete. {total - failed} successes, {failed} failures. Database saved.", "green")
        self.is_processing = False
        self.files_to_process = set() # Clear queue
        
        # Update GUI from thread safely works in CTk usually, but ideally use events
        self.btn_scan.configure(state="normal")
        self.btn_process.configure(text="2. Process New Audio (AI)")
        self.btn_sync.configure(state="normal")
        self.progress_bar.set(0)

    # --- STEP 3: SYNCING LOGIC (rsync) ---

    def start_sync_thread(self):
        user = self.pi_user_entry.get()
        ip = self.pi_ip_entry.get()
        dest = self.pi_dest_entry.get()

        if not user or not ip or not dest:
            self.log("Error: Please fill in all Pi Configuration fields.", "red")
            return

        self.btn_sync.configure(state="disabled", text="Syncing...")
        self.log("Starting Sync process... Ensure SSH keys are set up.")
        threading.Thread(target=self.sync_process, args=(user, ip, dest), daemon=True).start()

    def sync_process(self, user, ip, dest):
        remote_host = f"{user}@{ip}"
        
        # Ensure local path ends with slash for rsync source behavior
        source_path = self.local_music_path
        if not source_path.endswith(os.sep):
             source_path += os.sep

        # Ensure remote path ends with slash
        if not dest.endswith("/"):
             dest += "/"

        # Command 1: Sync Music Files (Delete on remote if deleted locally)
        # -a: archive mode (preserves times, permissions)
        # -v: verbose
        # --delete: delete extraneous files from dest dirs
        cmd_music = ["rsync", "-av", "--delete", source_path, f"{remote_host}:{dest}"]

        self.log(f"Executing Music Sync: {' '.join(cmd_music)}")
        try:
            # Using run instead of Popen to wait for completion
            result = subprocess.run(cmd_music, capture_output=True, text=True, check=True)
            self.log(result.stdout)
        except subprocess.CalledProcessError as e:
             self.log(f"RSYNC MUSIC ERROR: {e.stderr}", "red")
             self.btn_sync.configure(state="normal", text="3. Sync to Pi (rsync)")
             return
        except FileNotFoundError:
             self.log("ERROR: rsync not found. Is it installed on your system?", "red")
             self.btn_sync.configure(state="normal", text="3. Sync to Pi (rsync)")
             return

        # Command 2: Push the Library JSON separately to ensure it's the last thing updated
        source_json = os.path.join(self.local_music_path, LIBRARY_FILENAME)
        cmd_json = ["rsync", "-av", source_json, f"{remote_host}:{dest}"]
        self.log(f"Executing Database Sync: {' '.join(cmd_json)}")
        try:
            result = subprocess.run(cmd_json, capture_output=True, text=True, check=True)
            self.log(result.stdout)
            self.log("--- SYNC COMPLETE ---", "green")
        except subprocess.CalledProcessError as e:
             self.log(f"RSYNC DB ERROR: {e.stderr}", "red")

        self.btn_sync.configure(state="normal", text="3. Sync to Pi (rsync)")


if __name__ == "__main__":
    app = PiPodSyncApp()
    app.mainloop()
