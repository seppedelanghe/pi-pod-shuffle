from statics import RAW_DB_FILE, LIBRARY_FILE
from helpers import load_json

def cmd_visualize(args):
    """High-quality similarity visualization with explicit anchor selection."""
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
    raw_path = os.path.join(args.dir, RAW_DB_FILE)

    library = load_json(lib_path)
    raw = load_json(raw_path)

    if not library or not raw:
        print("Library is empty. Run 'process' first.")
        return

    paths = list(library.keys())
    vectors = np.array([library[p] for p in paths])

    if vectors.shape[1] < 2:
        print("Not enough dimensions to visualize.")
        return

    # -------------------------------------------------
    # 1. VARIANCE-WEIGHTED DISTANCE SPACE
    # -------------------------------------------------
    variances = np.var(vectors, axis=0)
    weights = np.sqrt(variances / np.sum(variances))
    Xw = vectors * weights

    distances = pairwise_distances(Xw, metric="cosine")

    # -------------------------------------------------
    # 2. FIND ANCHOR SONG
    # -------------------------------------------------
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

    # -------------------------------------------------
    # 3. SIMILAR & DISSIMILAR
    # -------------------------------------------------
    d = distances[anchor_idx]

    nearest = np.argsort(d)[1:6]
    furthest = np.argsort(d)[-5:][::-1]

    print("\n=== ANCHOR SONG ===")
    print(anchor_name)

    print("\n--- Top 5 MOST SIMILAR ---")
    for i in nearest:
        print(f"{d[i]:.3f}  |  {os.path.basename(paths[i])}")

    print("\n--- Top 5 MOST DISSIMILAR ---")
    for i in furthest:
        print(f"{d[i]:.3f}  |  {os.path.basename(paths[i])}")

    # -------------------------------------------------
    # 4. UMAP (VISUALIZATION ONLY)
    # -------------------------------------------------
    umap = UMAP(
        n_neighbors=min(15, len(Xw) - 1),
        min_dist=0.1,
        metric="euclidean",
        random_state=42,
    )
    embedding = umap.fit_transform(Xw)

    tempos = np.array([raw[p][0] for p in paths])

    # -------------------------------------------------
    # 5. PLOTTING
    # -------------------------------------------------
    fig, ax = plt.subplots(figsize=(10, 8))

    sc = ax.scatter(
        embedding[:, 0],
        embedding[:, 1],
        c=tempos,
        cmap="viridis",
        s=50,
        alpha=0.8,
    )

    # Anchor
    ax.scatter(
        embedding[anchor_idx, 0],
        embedding[anchor_idx, 1],
        color="red",
        s=160,
        label="Anchor",
        zorder=5,
    )

    # Nearest neighbors
    for i in nearest:
        ax.scatter(
            embedding[i, 0],
            embedding[i, 1],
            color="lime",
            s=120,
            zorder=4,
        )
        ax.plot(
            [embedding[anchor_idx, 0], embedding[i, 0]],
            [embedding[anchor_idx, 1], embedding[i, 1]],
            color="gray",
            alpha=0.4,
        )

    # Furthest neighbors
    for i in furthest:
        ax.scatter(
            embedding[i, 0],
            embedding[i, 1],
            color="black",
            s=80,
            alpha=0.7,
        )

    ax.set_title(f"Similarity Map (Anchor: {anchor_name})")
    ax.set_xlabel("UMAP-1")
    ax.set_ylabel("UMAP-2")
    ax.grid(alpha=0.2)

    cbar = plt.colorbar(sc, ax=ax)
    cbar.set_label("Tempo (BPM)")

    plt.tight_layout()
    plt.show()
