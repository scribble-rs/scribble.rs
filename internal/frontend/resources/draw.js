//Notice for core code of the floodfill, which has since then been heavily
//changed.
//Copyright(c) Max Irwin - 2011, 2015, 2016
//Repo: https://github.com/binarymax/floodfill.js
//MIT License

function floodfillData(data, x, y, fillcolor, width, height) {
    const length = data.length;
    let i = (x + y * width) * 4;

    //Fill coordinates are out of bounds
    if (i < 0 || i >= length) {
        return false;
    }

    //We check whether the target pixel is already the desired color, since
    //filling wouldn't change any of the pixels in this case.
    const targetcolor = [data[i], data[i + 1], data[i + 2]];
    if (
        targetcolor[0] === fillcolor.r &&
        targetcolor[1] === fillcolor.g &&
        targetcolor[2] === fillcolor.b) {
        return false;
    }

    let e = i, w = i, me, mw, w2 = width * 4;
    let j;

    //Previously we used Array.push and Array.pop here, with which the method
    //took between 70ms and 80ms on a rather strong machine with a FULL HD monitor.
    //Since Q can never be required to be bigger than the amount of maximum
    //pixels (width*height), we preallocate Q with that size. While not all of
    //the space might be needed, this is cheaper than reallocating multiple times.
    //This improved the time from 70ms-80ms to 50ms-60ms.
    const Q = new Array(width * height);
    let nextQIndex = 0;
    Q[nextQIndex++] = i;

    while (nextQIndex > 0) {
        i = Q[--nextQIndex];
        if (pixelCompareAndSet(i, targetcolor, fillcolor, data)) {
            e = i;
            w = i;
            mw = Math.floor(i / w2) * w2; //left bound
            me = mw + w2;               //right bound
            while (mw < w && mw <= (w -= 4) && pixelCompareAndSet(w, targetcolor, fillcolor, data)); //go left until edge hit
            while (me > e && me > (e += 4) && pixelCompareAndSet(e, targetcolor, fillcolor, data)); //go right until edge hit
            for (j = w + 4; j < e; j += 4) {
                if (j - w2 >= 0 && pixelCompare(j - w2, targetcolor, data)) Q[nextQIndex++] = j - w2; //queue y-1
                if (j + w2 < length && pixelCompare(j + w2, targetcolor, data)) Q[nextQIndex++] = j + w2; //queue y+1
            }
        }
    }

    return data;
};

function pixelCompare(i, targetcolor, data) {
    return (
        targetcolor[0] === data[i] &&
        targetcolor[1] === data[i + 1] &&
        targetcolor[2] === data[i + 2]
    );
};

function pixelCompareAndSet(i, targetcolor, fillcolor, data) {
    if (pixelCompare(i, targetcolor, data)) {
        data[i] = fillcolor.r;
        data[i + 1] = fillcolor.g;
        data[i + 2] = fillcolor.b;
        return true;
    }
    return false;
};

function floodfillUint8ClampedArray(data, x, y, color, width, height) {
    if (isNaN(width) || width < 1) throw new Error("argument 'width' must be a positive integer");
    if (isNaN(height) || height < 1) throw new Error("argument 'height' must be a positive integer");
    if (isNaN(x) || x < 0) throw new Error("argument 'x' must be a positive integer");
    if (isNaN(y) || y < 0) throw new Error("argument 'y' must be a positive integer");
    if (width * height * 4 !== data.length) throw new Error("width and height do not fit Uint8ClampedArray dimensions");

    const xi = Math.floor(x);
    const yi = Math.floor(y);

    return floodfillData(data, xi, yi, color, width, height);
};

// Code for line drawing, not related to the floodfill repo.
// Hence it's all BSD licensed.

function drawLine(context, imageData, x1, y1, x2, y2, color, width) {
    const coords = prepareDrawLineCoords(context, x1, y1, x2, y2, width);
    _drawLineNoPut(imageData, coords, color, width);
    context.putImageData(imageData, 0, 0, 0, 0, coords.right, coords.bottom);
};

// This implementation directly access the canvas data and does not
// put it back into the canvas context directly. This saved us not
// only from calling put, which is relatively cheap, but also from
// calling getImageData all the time.
function drawLineNoPut(context, imageData, x1, y1, x2, y2, color, width) {
    _drawLineNoPut(imageData, prepareDrawLineCoords(context, x1, y1, x2, y2, width), color, width);
};

function _drawLineNoPut(imageData, coords, color, width) {
    const { x1, y1, x2, y2, left, top, right, bottom } = coords;

    // off canvas, so don't draw anything
    if (right - left === 0 || bottom - top === 0) {
        return;
    }

    const circleMap = generateCircleMap(Math.floor(width / 2));
    const offset = Math.floor(circleMap.length / 2);

    for (let ix = 0; ix < circleMap.length; ix++) {
        for (let iy = 0; iy < circleMap[ix].length; iy++) {
            if (circleMap[ix][iy] === 1 || (x1 === x2 && y1 === y2 && circleMap[ix][iy] === 2)) {
                const newX1 = x1 + ix - offset;
                const newY1 = y1 + iy - offset;
                const newX2 = x2 + ix - offset;
                const newY2 = y2 + iy - offset;
                drawBresenhamLine(imageData, newX1, newY1, newX2, newY2, color);
            }
        }
    }
}

function prepareDrawLineCoords(context, x1, y1, x2, y2, width) {
    // the coordinates must be whole numbers to improve performance.
    // also, decimals as coordinates is not making sense.
    x1 = Math.floor(x1);
    y1 = Math.floor(y1);
    x2 = Math.floor(x2);
    y2 = Math.floor(y2);

    // calculate bounding box
    const left = Math.max(0, Math.min(context.canvas.width, Math.min(x1, x2) - width));
    const top = Math.max(0, Math.min(context.canvas.height, Math.min(y1, y2) - width));
    const right = Math.max(0, Math.min(context.canvas.width, Math.max(x1, x2) + width));
    const bottom = Math.max(0, Math.min(context.canvas.height, Math.max(y1, y2) + width));

    return {
        x1: x1,
        y1: y1,
        x2: x2,
        y2: y2,
        left: left,
        top: top,
        right: right,
        bottom: bottom,
    };
}
function drawBresenhamLine(imageData, x1, y1, x2, y2, color) {
    const dx = Math.abs(x2 - x1);
    const dy = Math.abs(y2 - y1);
    const sx = (x1 < x2) ? 1 : -1;
    const sy = (y1 < y2) ? 1 : -1;
    let err = dx - dy;

    while (true) {
        //check if pixel is inside the canvas
        if (!(x1 < 0 || x1 >= imageData.width || y1 < 0 || y1 >= imageData.height)) {
            setPixel(imageData, x1, y1, color);
        }

        if ((x1 === x2) && (y1 === y2)) break;
        const e2 = 2 * err;
        if (e2 > -dy) {
            err -= dy;
            x1 += sx;
        }
        if (e2 < dx) {
            err += dx;
            y1 += sy;
        }
    }
}

// We cache them, as we need quite a lot of them, but the pencil size usually
// doesn't change that often. There's also not many sizes, so we don't need to
// worry about invalidating anything.
let cachedCircleMaps = {};

function generateCircleMap(radius) {
    const cached = cachedCircleMaps[radius];
    if (cached) {
        return cached;
    }

    const diameter = 2 * radius;
    const circleData = new Array(diameter);

    for (let x = 0; x < diameter; x++) {
        circleData[x] = new Array(diameter);
        for (let y = 0; y < diameter; y++) {
            const distanceToRadius = Math.sqrt(Math.pow(radius - x, 2) + Math.pow(radius - y, 2));
            if (distanceToRadius > radius) {
                circleData[x][y] = 0;
            } else if (distanceToRadius < radius - 2) {
                circleData[x][y] = 2;
            } else {
                circleData[x][y] = 1;
            }
        }
    }

    cachedCircleMaps[radius] = circleData;
    return circleData;
}

function setPixel(imageData, x, y, color) {
    const offset = (y * imageData.width + x) * 4;
    imageData.data[offset] = color.r;
    imageData.data[offset + 1] = color.g;
    imageData.data[offset + 2] = color.b;
}

//We accept both #RRGGBB and RRGGBB. Both are treated case insensitive.
function hexStringToRgbColorObject(hexString) {
    if (!hexString) {
        return { r: 0, g: 0, b: 0 };
    }
    const hexColorsRegex = /#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})/i;
    const match = hexString.match(hexColorsRegex)
    return { r: parseInt(match[1], 16), g: parseInt(match[2], 16), b: parseInt(match[3], 16) };
}

const colorMap = [
    { hex: '#ffffff', rgb: hexStringToRgbColorObject('#ffffff') },
    { hex: '#c1c1c1', rgb: hexStringToRgbColorObject('#c1c1c1') },
    { hex: '#ef130b', rgb: hexStringToRgbColorObject('#ef130b') },
    { hex: '#ff7100', rgb: hexStringToRgbColorObject('#ff7100') },
    { hex: '#ffe400', rgb: hexStringToRgbColorObject('#ffe400') },
    { hex: '#00cc00', rgb: hexStringToRgbColorObject('#00cc00') },
    { hex: '#00b2ff', rgb: hexStringToRgbColorObject('#00b2ff') },
    { hex: '#231fd3', rgb: hexStringToRgbColorObject('#231fd3') },
    { hex: '#a300ba', rgb: hexStringToRgbColorObject('#a300ba') },
    { hex: '#d37caa', rgb: hexStringToRgbColorObject('#d37caa') },
    { hex: '#a0522d', rgb: hexStringToRgbColorObject('#a0522d') },
    { hex: '#592f2a', rgb: hexStringToRgbColorObject('#592f2a') },
    { hex: '#ecbcb4', rgb: hexStringToRgbColorObject('#ecbcb4') },
    { hex: '#000000', rgb: hexStringToRgbColorObject('#000000') },
    { hex: '#4c4c4c', rgb: hexStringToRgbColorObject('#4c4c4c') },
    { hex: '#740b07', rgb: hexStringToRgbColorObject('#740b07') },
    { hex: '#c23800', rgb: hexStringToRgbColorObject('#c23800') },
    { hex: '#e8a200', rgb: hexStringToRgbColorObject('#e8a200') },
    { hex: '#005510', rgb: hexStringToRgbColorObject('#005510') },
    { hex: '#00569e', rgb: hexStringToRgbColorObject('#00569e') },
    { hex: '#0e0865', rgb: hexStringToRgbColorObject('#0e0865') },
    { hex: '#550069', rgb: hexStringToRgbColorObject('#550069') },
    { hex: '#a75574', rgb: hexStringToRgbColorObject('#a75574') },
    { hex: '#63300d', rgb: hexStringToRgbColorObject('#63300d') },
    { hex: '#492f31', rgb: hexStringToRgbColorObject('#492f31') },
    { hex: '#d1a3a4', rgb: hexStringToRgbColorObject('#d1a3a4') }
];

function indexToHexColor(index) {
    return colorMap[index].hex;
}

function indexToRgbColor(index) {
    return colorMap[index].rgb;
}

