var floodfill = (function() {

    //Copyright(c) Max Irwin - 2011, 2015, 2016
    //Repo: https://github.com/binarymax/floodfill.js
    //MIT License

    function floodfill(data,x,y,fillcolor,width,height) {
        var length = data.length;
        var i = (x+y*width)*4;

        //Fill coordinates are out of bounds
        if (i<0||i>=length) {
            return false;
        }

        var targetcolor = [data[i],data[i+1],data[i+2]];

        //We check whether the target pixel is already the desired color, since
        //filling wouldn't change any of the pixels in this case.
        if (
            targetcolor[0] === fillcolor.r &&
            targetcolor[1] === fillcolor.g &&
            targetcolor[2] === fillcolor.b) {
            return false;
        }

        var e = i, w = i, me, mw, w2 = width*4;
        var j;

        //Previously we used Array.push and Array.pop here, with which the method
        //took between 70ms and 80ms on a rather strong machine with a FULL HD monitor.
        //Since Q can never be required to be bigger than the amount of maximum
        //pixels (width*height), we preallocate Q with that size. While not all of
        //the space might be needed, this is cheaper than reallocating multiple times.
        //This improved the time from 70ms-80ms to 50ms-60ms.
        var Q = new Array(width*height);
        var nextQIndex = 0;
        Q[nextQIndex++] = i;

        while(nextQIndex > 0) {
            i = Q[--nextQIndex];
            if(pixelCompareAndSet(i,targetcolor,fillcolor,data,length)) {
                e = i;
                w = i;
                mw = Math.floor(i/w2)*w2; //left bound
                me = mw+w2;               //right bound
                while(mw<w && mw<=(w-=4) && pixelCompareAndSet(w,targetcolor,fillcolor,data)); //go left until edge hit
                while(me>e && me>(e+=4) && pixelCompareAndSet(e,targetcolor,fillcolor,data)); //go right until edge hit
                for(j=w+4;j<e;j+=4) {
                    if(j-w2>=0     && pixelCompare(j-w2,targetcolor,data)) Q[nextQIndex++]=j-w2; //queue y-1
                    if(j+w2<length && pixelCompare(j+w2,targetcolor,data)) Q[nextQIndex++]=j+w2; //queue y+1
                }
            }
        }

        return data;
    };

    function pixelCompare(i,targetcolor,data) {
        return (
            targetcolor[0] === data[i] &&
            targetcolor[1] === data[i+1] &&
            targetcolor[2] === data[i+2]
        );
    };

    function pixelCompareAndSet(i,targetcolor,fillcolor,data) {
        if(pixelCompare(i,targetcolor,data)) {
            //fill the color
            data[i]   = fillcolor.r;
            data[i+1] = fillcolor.g;
            data[i+2] = fillcolor.b;
            return true;
        }
        return false;
    };

    function fillUint8ClampedArray(data,x,y,color,width,height) {
        if (!data instanceof Uint8ClampedArray) throw new Error("data must be an instance of Uint8ClampedArray");
        if (isNaN(width)  || width<1)  throw new Error("argument 'width' must be a positive integer");
        if (isNaN(height) || height<1) throw new Error("argument 'height' must be a positive integer");
        if (isNaN(x) || x<0) throw new Error("argument 'x' must be a positive integer");
        if (isNaN(y) || y<0) throw new Error("argument 'y' must be a positive integer");
        if (width*height*4!==data.length) throw new Error("width and height do not fit Uint8ClampedArray dimensions");

        var xi = Math.floor(x);
        var yi = Math.floor(y);

        return floodfill(data,xi,yi,color,width,height);
    };

    function fillContext(x,y,color) {
        var ctx = this;

        var image = ctx.getImageData(0,0,ctx.canvas.width,ctx.canvas.height);
        var width = image.width;
        var height = image.height;
        
        if(width>0 && height>0) {
            const hasFilled = fillUint8ClampedArray(image.data,x,y,color,width,height);
            if (hasFilled) {
                ctx.putImageData(image,0,0);
            }
            return hasFilled;
        }

        return false;
    };

    if (typeof CanvasRenderingContext2D != 'undefined') {
        CanvasRenderingContext2D.prototype.fillFlood = fillContext;
    };

    return fillUint8ClampedArray;
})();
