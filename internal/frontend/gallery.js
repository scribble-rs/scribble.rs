document.getElementById("prev").addEventListener("click", () => {
    prevDrawing();
});

document.getElementById("next").addEventListener("click", () => {
    nextDrawing();
});

/**
 * @returns {Promise<IDBDatabase>}
 */
const openDB = () => {
    const db = indexedDB.open("scribblers", 1);

    db.onupgradeneeded = (event) => {
        const db = event.target.result;
        const objectStore = db.createObjectStore("gallery", { keyPath: "id" });
        // No index, as we store an array.
    };

    return new Promise((resolve, reject) => {
        db.onsuccess = () => {
            resolve(db.result);
        };
        db.onerror = () => {
            reject(db.error);
        };
    });
};

const dbPromise = openDB();

const getGalleryEntry = async (store, id) => {
    return new Promise((resolve, reject) => {
        const gallery = store.get(id);
        gallery.onsuccess = (event) => {
            const galleryData = event.target.result;
            resolve(galleryData);
        };
        gallery.onerror = () => {
            reject(gallery.error);
        };
    });
};

const getGallery = () => {
    return new Promise(async (resolve, reject) => {
        const db = await dbPromise;
        const store = db.transaction("gallery").objectStore("gallery");
        const cachedGallery = await getGalleryEntry(store, "{{.LobbyID}}");

        fetch(
            "{{.RootPath}}/v1/lobby/{{.LobbyID}}/gallery?" +
                new URLSearchParams({
                    local_cache_count: cachedGallery
                        ? cachedGallery.data.length
                        : 0,
                }).toString(),
        )
            .then((response) => {
                if (response.status === 204) {
                    console.log(
                        "No new gallery data for lobby {{.LobbyID}} available.",
                    );
                    resolve(cachedGallery ? cachedGallery.data : []);
                    return;
                }

                if (response.status === 200) {
                    response
                        .json()
                        .then((json) => {
                            const store = db
                                .transaction("gallery", "readwrite")
                                .objectStore("gallery");
                            store.put({
                                id: "{{.LobbyID}}",
                                data: json,
                            });
                            console.log(
                                "Latest gallery for lobby {{.LobbyID}} stored.",
                            );
                            return json;
                        })
                        .then(resolve);
                    return;
                }

                console.log("Unknown error, falling back to cached value");
                resolve(cachedGallery.data);
            })
            .catch((err) => {
                if (cachedGallery && cachedGallery.data.length > 0) {
                    resolve(cachedGallery.data);
                } else {
                    reject(err);
                }
            });
    });
};

const word = document.getElementById("word");

const drawingBoard = document.getElementById("drawing-board");
const context = drawingBoard.getContext("2d", { alpha: false });
let imageData;

function clear(context) {
    context.fillStyle = "#FFFFFF";
    context.fillRect(0, 0, drawingBoard.width, drawingBoard.height);
    // Refetch, as we don't manually fill here.
    imageData = context.getImageData(
        0,
        0,
        context.canvas.width,
        context.canvas.height,
    );
}
clear(context);

function setDrawing(drawing) {
    clear(context);

    word.innerText = drawing.word;

    drawing.events.forEach((drawElement) => {
        const drawData = drawElement.data;
        if (drawElement.type === "fill") {
            floodfillUint8ClampedArray(
                imageData.data,
                drawData.x,
                drawData.y,
                indexToRgbColor(drawData.color),
                imageData.width,
                imageData.height,
            );
        } else if (drawElement.type === "line") {
            drawLineNoPut(
                context,
                imageData,
                drawData.x,
                drawData.y,
                drawData.x2,
                drawData.y2,
                indexToRgbColor(drawData.color),
                drawData.width,
            );
        } else {
            console.log("Unknown draw element type: " + drawData.type);
        }
    });

    context.putImageData(imageData, 0, 0);
}

let currentIndex = 0;
let galleryData;

getGallery().then((data) => {
    if (data.length > 0) {
        setDrawing(data[0]);
    }
    galleryData = data;
});

function prevDrawing() {
    if (!galleryData) {
        return;
    }

    if (currentIndex <= 0) {
        return;
    }

    currentIndex = currentIndex - 1;
    setDrawing(galleryData[currentIndex]);
}

function nextDrawing() {
    if (!galleryData) {
        return;
    }

    if (currentIndex >= galleryData.length - 1) {
        return;
    }

    currentIndex = currentIndex + 1;
    setDrawing(galleryData[currentIndex]);
}

dbPromise.then((db) => {
    const galleryStore = db.transaction("gallery").objectStore("gallery");
    const galleryCursor = galleryStore.openCursor();
    galleryCursor.onsuccess = async (event) => {
        const cursor = event.target.result;
        if (cursor) {
            const entry = await cursor.value;
            console.log(entry.id);
            cursor.continue();
        }
    };
});
