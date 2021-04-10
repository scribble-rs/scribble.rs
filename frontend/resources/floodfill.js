var floodfill = (function() {

    //Copyright(c) Max Irwin - 2011, 2015, 2016
    //Repo: https://github.com/binarymax/floodfill.js
    //MIT License

    function floodfill(data,x,y,fillcolor,width,height) {

        var length = data.length;
        var i = (x+y*width)*4;
        var targetcolor = [data[i],data[i+1],data[i+2],data[i+3]];
        if(!pixelCompare(i,targetcolor,fillcolor,data,length)) { return false; }

        var e = i, w = i, me, mw, w2 = width*4;
        var j;

        //Previously we used Array.push and Array.pop here. The method took between 70ms and 80ms
        //on a rather strong machine with a FULL HD monitor.
        //Observastions show that this stack grows up to 1_000_000 in this scenario.
        //Therefore we allocate even more than that upfront to avoid reallocation.
        //This improved the time from 70ms-80ms to 50ms-60ms.

        //FIXME: Another optimization would be to take the canvas size into account and make estimations.
        //This changes won't be upstreamed to the mainrepo yet, as they are very specific to our usecase.
        //If the canvas size estimation is applied, this would be fine to upstream I guess.
        var Q = new Array(1500000);
        var nextQIndex = 0;
        Q[nextQIndex++] = i;

        while(nextQIndex > 0) {
            i = Q[--nextQIndex];
            if(pixelCompareAndSet(i,targetcolor,fillcolor,data,length)) {
                e = i;
                w = i;
                mw = parseInt(i/w2)*w2; //left bound
                me = mw+w2;             //right bound
                while(mw<w && mw<(w-=4) && pixelCompareAndSet(w,targetcolor,fillcolor,data,length)); //go left until edge hit
                while(me>e && me>(e+=4) && pixelCompareAndSet(e,targetcolor,fillcolor,data,length)); //go right until edge hit
                for(j=w+4;j<e;j+=4) {
                    if(j-w2>=0     && pixelCompare(j-w2,targetcolor,fillcolor,data,length)) Q[nextQIndex++]=j-w2; //queue y-1
                    if(j+w2<length && pixelCompare(j+w2,targetcolor,fillcolor,data,length)) Q[nextQIndex++]=j+w2; //queue y+1
                }
            }
        }
        return data;
    };

    function pixelCompare(i,targetcolor,fillcolor,data,length) {
        if (i<0||i>=length) return false; //out of bounds

        if (
            targetcolor[0] === fillcolor.r &&
            targetcolor[1] === fillcolor.g &&
            targetcolor[2] === fillcolor.b
        ) return false; //target is same as fill

        if (
            targetcolor[0] === data[i] &&
            targetcolor[1] === data[i+1] &&
            targetcolor[2] === data[i+2]
        ) return true; //target matches surface

        return false; //no match
    };

    function pixelCompareAndSet(i,targetcolor,fillcolor,data,length) {
        if(pixelCompare(i,targetcolor,fillcolor,data,length)) {
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

    var getComputedColor = function(c) {
        var temp = document.createElement("div");
        var color = {r:0,g:0,b:0};
        temp.style.color = c;
        temp.style.display = "none";
        document.body.appendChild(temp);
        //Use native window.getComputedStyle to parse any CSS color pattern
        var style = window.getComputedStyle(temp,null).color;
        document.body.removeChild(temp);

        var recol = /([\.\d]+)/g;
        var vals  = style.match(recol);
        if (vals && vals.length>2) {
            //Coerce the string value into an rgba object
            color.r = parseInt(vals[0])||0;
            color.g = parseInt(vals[1])||0;
            color.b = parseInt(vals[2])||0;
        }
        return color;
    };

    function fillContext(x,y,left,top,right,bottom) {
        var ctx  = this;

        //Gets the rgba color from the context fillStyle
        var color = getComputedColor(this.fillStyle);

        //Defaults and type checks for image boundaries
        left     = (isNaN(left)) ? 0 : left;
        top      = (isNaN(top)) ? 0 : top;
        right    = (!isNaN(right)&&right) ? Math.min(Math.abs(right),ctx.canvas.width) : ctx.canvas.width;
        bottom   = (!isNaN(bottom)&&bottom) ? Math.min(Math.abs(bottom),ctx.canvas.height) : ctx.canvas.height;

        var image = ctx.getImageData(left,top,right,bottom);

        var data = image.data;
        var width = image.width;
        var height = image.height;

        if(width>0 && height>0) {
            const hasFilled = fillUint8ClampedArray(data,x,y,color,width,height);
            if (hasFilled) {
                ctx.putImageData(image,left,top);
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
