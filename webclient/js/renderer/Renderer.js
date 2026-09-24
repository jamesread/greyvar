/**
 *
import { renderGrid } from './rendererGrid.js'
 
  onCanvasResized() {
    return

    // Uset offset* values here, as .width and .height are set to whatever is in the initial css.
    this.width = this.canvas.offsetWidth
    this.height = this.canvas.offsetHeight

    let l = Math.min(this.width, this.height)

    this.canvas.width = l;
    this.canvas.height = l;

    this.drawScale = l / 256; // Greyvar uses a logical 256 x 256 resolution, and the renderers "upscale" as needed

    return this.drawScale
  }


