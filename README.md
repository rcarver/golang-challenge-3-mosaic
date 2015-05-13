# golang challenge 3

This is an official entry.

* [Challenge](http://golang-challenge.com/go-challenge3/)
* [Mosaic](http://en.wikipedia.org/wiki/Photographic_mosaic)
* [go-colorful](https://github.com/lucasb-eyer/go-colorful)


# Algorithm

  * Build color palette of Len N from images on hand
  * Take Original image
  * Grid = Convert Original to grid
  * PatternImage = Draw Grid as pixels with FloydSteinberg and palette (dither)
  * Output = new image as pixels * unit size
  * For each pixel in PatternImage, pull an image from the palette and place it at x,y
  * http://tech-algorithm.com/articles/nearest-neighbor-image-scaling/

