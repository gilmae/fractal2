# Fractals

A code base for generating fractal images.

This may replace the existing [mandelbrot](https://github.com/gilmae/mandelbrot)
program. Well... the point is more to try and take the lessons of the previous
application, and remove some of the cruft that grew on it. 

1. I think it was a mistake to make the fractal libray and then have mandelbrot depend on it. It introduced so much complexity when I was trying, and failing, to introduce the BigFloat versions. Instead I think mandelbrot vs julia vs burning ship could be function/plugins.
2. I want to remove some of the in memory data structures I had, the ones that held plotted points in a hash until after all points had been generated. If we can get the pre-ploted points to come out of a channel, be calculated, be put into another channel, and then come out of _that_ channel and be set directly into an image, that would save a lot of memory space while running. I think.
