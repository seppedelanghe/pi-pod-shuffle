from statics import LIBRARY_FILE
from helpers import load_json

def cmd_visualize(args):
    """
    Visualizes similarity using RAW embeddings for calculation 
    and UMAP for the 2D plot.
    """
    try:
        import matplotlib.pyplot as plt
        from sklearn.metrics import pairwise_distances
        from umap import UMAP
        import numpy as np
        import os
    except ImportError as e:
        print(f"Error: Missing visualization dependency ({e})")
        print("pip install matplotlib umap-learn scikit-learn")
        return

    lib_path = os.path.join(args.dir, LIBRARY_FILE)
    library = load_json(lib_path)

    if not library:
        print("Library is empty. Run 'process' first.")
        return

    # 1. LOAD RAW EMBEDDINGS
    # We maintain the order of paths to map indices back to filenames
    paths = list(library.keys())
    
    # Create the matrix (N_songs x 512)
    # This IS the raw data from PANNs
    vectors = np.array([library[p] for p in paths])

    print(f"Loaded {len(vectors)} songs with {vectors.shape[1]} dimensions.")

    # 2. CALCULATE SIMILARITY ON RAW DATA
    # We use Cosine Distance (1 - Cosine Similarity)
    # This happens in 512-dim space for maximum accuracy
    distances = pairwise_distances(vectors, metric="cosine")

    # 3. SELECT ANCHOR SONG
    if args.anchor:
        matches = [
            i for i, p in enumerate(paths)
            if args.anchor.lower() in os.path.basename(p).lower()
        ]
        if not matches:
            print(f"No match found for anchor '{args.anchor}'")
            return
        anchor_idx = matches[0]
    else:
        anchor_idx = np.random.randint(len(paths))

    anchor_path = paths[anchor_idx]
    anchor_name = os.path.basename(anchor_path)

    # 4. FIND NEIGHBORS (MATH STEP)
    # We look at the row in the distance matrix corresponding to our anchor
    d_scores = distances[anchor_idx]

    # argsort gives us the indices of the sorted array
    # [1:6] skips index 0 (which is the song itself, distance 0.0)
    nearest_indices = np.argsort(d_scores)[1:6]
    
    # [-5:][::-1] gets the last 5 (highest distance) and reverses them
    furthest_indices = np.argsort(d_scores)[-5:][::-1]

    print("\n" + "="*40)
    print(f"ANCHOR: {anchor_name}")
    print("="*40)

    print(f"{'DIST':<8} | {'TOP 5 SIMILAR'}")
    print("-" * 40)
    for i in nearest_indices:
        print(f"{d_scores[i]:.4f}   | {os.path.basename(paths[i])}")

    print("\n" + "-" * 40)
    print(f"{'DIST':<8} | {'TOP 5 DISSIMILAR'}")
    print("-" * 40)
    for i in furthest_indices:
        print(f"{d_scores[i]:.4f}   | {os.path.basename(paths[i])}")

    # 5. REDUCE DIMENSIONS (VISUALIZATION STEP ONLY)
    # We reduce 512 -> 2 just for the graph. 
    # This does NOT affect the text results printed above.
    print("\nCalculating 2D projection for plot...")
    reducer = UMAP(
        n_neighbors=15, 
        min_dist=0.1, 
        metric='cosine', # It's good practice to match the metric used above
        random_state=42
    )
    embedding_2d = reducer.fit_transform(vectors)

    # 6. PLOT
    fig, ax = plt.subplots(figsize=(10, 8))

    # Plot all songs as faint blue dots
    ax.scatter(
        embedding_2d[:, 0],
        embedding_2d[:, 1],
        c='lightgray',
        s=30,
        alpha=0.5,
        label='Library'
    )

    # Plot Anchor (Red Star)
    ax.scatter(
        embedding_2d[anchor_idx, 0],
        embedding_2d[anchor_idx, 1],
        c='red',
        s=200,
        marker='*',
        label='Anchor',
        zorder=10
    )

    # Plot Nearest (Green Circles)
    # connect them with lines to the anchor
    for i in nearest_indices:
        ax.scatter(
            embedding_2d[i, 0],
            embedding_2d[i, 1],
            c='green',
            s=100,
            zorder=9
        )
        # Draw line
        ax.plot(
            [embedding_2d[anchor_idx, 0], embedding_2d[i, 0]],
            [embedding_2d[anchor_idx, 1], embedding_2d[i, 1]],
            c='green',
            alpha=0.3
        )
        # Add simple text label
        ax.text(
            embedding_2d[i, 0], 
            embedding_2d[i, 1], 
            os.path.basename(paths[i])[:15], # truncate long names
            fontsize=8
        )

    ax.legend()
    ax.set_title(f"Similarity Space: {anchor_name}")
    ax.set_xlabel("UMAP Dimension 1")
    ax.set_ylabel("UMAP Dimension 2")
    
    plt.tight_layout()
    plt.show()
