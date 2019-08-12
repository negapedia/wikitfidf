from Cython.Distutils import build_ext
from distutils.core import setup
from distutils.extension import Extension

# cython: language_level=3
ext_modules = [
    Extension("StemStopwClean", ["DeStemmer.py"])
    #   ... all your modules that need be compiled ...
]
setup(
    name='DeStemmer',
    cmdclass={'build_ext': build_ext},
    ext_modules=ext_modules,
    language_level=3
)
