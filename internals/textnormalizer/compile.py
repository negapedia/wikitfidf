from distutils.core import setup
from distutils.extension import Extension

from Cython.Distutils import build_ext

# cython: language_level=3
ext_modules = [
    Extension("StemStopwClean", ["StopwordsCleaner_Stemming.py"])
    #   ... all your modules that need be compiled ...
]
setup(
    name='StemStopwClean',
    cmdclass={'build_ext': build_ext},
    ext_modules=ext_modules,
    language_level=3
)
