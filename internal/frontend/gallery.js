document.getElementById("prev").addEventListener("click", () => {
    prevDrawing();
});

document.getElementById("next").addEventListener("click", () => {
    nextDrawing();
});

const getGallery = () => {
    return new Promise((resolve, reject) => {
        const cachedGallery = sessionStorage.getItem("cached_gallery");
        if (cachedGallery) {
            resolve(JSON.parse(cachedGallery));
            return;
        }

        fetch("{{.RootPath}}/v1/lobby/{{.LobbyID}}/gallery")
            .then((response) => {
                response
                    .json()
                    .then((json) => {
                        sessionStorage.setItem(
                            "cached_gallery",
                            JSON.stringify(json),
                        );
                        return json;
                    })
                    .then(resolve);
            })
            .catch(reject);
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
    setDrawing(data[0]);
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
