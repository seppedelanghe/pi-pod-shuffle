import os
import argparse
from openai import OpenAI

def get_corrected_filename(client, model, original_path):
    """Sends the filename to the selected model for formatting."""
    filename = os.path.basename(original_path)
    
    prompt = (
        f"Extract the track ID and track name from this filename: '{filename}'. "
        f"Return ONLY the corrected filename in the format: <track-id> - <track-name>.flac. "
        f"If no track ID is found, use '00'. Do not include explanations."
    )

    response = client.chat.completions.create(
        model=model,
        messages=[
            {"role": "system", "content": "You are a precise music library assistant."},
            {"role": "user", "content": prompt}
        ]
    )
    
    return response.choices[0].message.content.strip()

def main():
    parser = argparse.ArgumentParser(description="Rename FLAC files using OpenAI or LM Studio.")
    parser.add_argument("path", help="The root directory to search for FLAC files")
    parser.add_argument("--model", default="gpt-4o-mini", help="Model name (e.g., gpt-4o-mini or your local model ID)")
    parser.add_argument("--local", action="store_true", help="Use local LM Studio server instead of OpenAI")
    parser.add_argument("--url", default="http://localhost:1234/v1", help="Local server URL (default: http://localhost:1234/v1)")
    parser.add_argument("--dry-run", action="store_true", help="Preview changes without a prompt")
    
    args = parser.parse_args()

    # Client Configuration
    if args.local:
        print(f"Connecting to local LM Studio server at {args.url}...")
        client = OpenAI(base_url=args.url, api_key="lm-studio") # Dummy key for local
    else:
        print("Connecting to OpenAI API...")
        client = OpenAI() # Uses OPENAI_API_KEY environment variable

    if not os.path.isdir(args.path):
        print(f"Error: {args.path} is not a valid directory.")
        return

    pending_changes = []
    print(f"Scanning files in: {args.path}...")

    # 1. Collect all proposed changes
    for root, _, files in os.walk(args.path):
        for filename in files:
            if filename.lower().endswith(".flac"):
                old_path = os.path.join(root, filename)
                
                try:
                    print(f" Consulting model for: {filename}")
                    new_filename = get_corrected_filename(client, args.model, old_path)
                    
                    if filename != new_filename:
                        pending_changes.append((old_path, new_filename))
                    else:
                        print("  [SKIPPED] Name is already correct.")
                        
                except Exception as e:
                    print(f" [ERROR] Could not process {filename}: {e}")

    # 2. Review Phase
    if not pending_changes:
        print("\nNo files need renaming.")
        return

    print("\n--- PROPOSED CHANGES ---")
    for old_p, new_f in pending_changes:
        print(f"FROM: {os.path.basename(old_p)}")
        print(f"TO:   {new_f}")

    if args.dry_run:
        print(f"\nDry run complete. {len(pending_changes)} changes found.")
        return

    # 3. Confirmation Phase
    print(f"\nTotal changes: {len(pending_changes)}")
    confirm = input("Proceed with renaming? (y/N): ").lower()
    
    if confirm == 'y':
        for old_path, new_filename in pending_changes:
            new_path = os.path.join(os.path.dirname(old_path), new_filename)
            try:
                os.rename(old_path, new_path)
                print(f"[SUCCESS] Renamed: {new_filename}")
            except Exception as e:
                print(f"[ERROR] Failed to rename {old_path}: {e}")
        print("\nRenaming complete.")
    else:
        print("\nOperation cancelled. No files were changed.")

if __name__ == "__main__":
    main()
