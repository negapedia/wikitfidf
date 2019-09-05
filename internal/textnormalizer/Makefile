# Makefile for creating our standalone Cython program
# Shamefully copied from https://github.com/cython/cython/blob/master/Demos/embed/Makefile

PYTHON := python3
PYVERSION := $(shell $(PYTHON) -c "import sys; print(sys.version[:3])")
PYPREFIX := $(shell $(PYTHON) -c "import sys; print(sys.prefix)")

INCDIR := $(shell $(PYTHON) -c "from distutils import sysconfig; print(sysconfig.get_python_inc())")
PLATINCDIR := $(shell $(PYTHON) -c "from distutils import sysconfig; print(sysconfig.get_python_inc(plat_specific=True))")
LIBDIR1 := $(shell $(PYTHON) -c "from distutils import sysconfig; print(sysconfig.get_config_var('LIBDIR'))")
LIBDIR2 := $(shell $(PYTHON) -c "from distutils import sysconfig; print(sysconfig.get_config_var('LIBPL'))")
PYLIB := $(shell $(PYTHON) -c "from distutils import sysconfig; print(sysconfig.get_config_var('LIBRARY')[3:-2])")

CC := $(shell $(PYTHON) -c "import distutils.sysconfig; print(distutils.sysconfig.get_config_var('CC'))")
LINKCC := $(shell $(PYTHON) -c "import distutils.sysconfig; print(distutils.sysconfig.get_config_var('LINKCC'))")
LINKFORSHARED := $(shell $(PYTHON) -c "import distutils.sysconfig; print(distutils.sysconfig.get_config_var('LINKFORSHARED'))")
LIBS := $(shell $(PYTHON) -c "import distutils.sysconfig; print(distutils.sysconfig.get_config_var('LIBS'))")
SYSLIBS :=  $(shell $(PYTHON) -c "import distutils.sysconfig; print(distutils.sysconfig.get_config_var('SYSLIBS'))")

textnormalizer: textnormalizer.o
	$(LINKCC) -o $@ $^ -L$(LIBDIR1) -L$(LIBDIR2) -l$(PYLIB) $(LIBS) $(SYSLIBS) $(LINKFORSHARED)

textnormalizer.o: textnormalizer.c
	$(CC) -c $^ -I$(INCDIR) -I$(PLATINCDIR)

textnormalizer.c: textnormalizer.pyx
	@cython --embed textnormalizer.pyx

nltk-master/:
	curl https://codeload.github.com/negapedia/nltk/zip/master | jar xv

all: nltk-master/ WikipediaMarkupCleaner.jar textnormalizer

clean:
	@rm -f textnormalizer.o textnormalizer.c

run: clean all 
	java -jar WikipediaMarkupCleaner.jar $(RESULTDIR)
	NLTK_DATA="nltk-master" PYTHONHOME=$(PYPREFIX) LD_LIBRARY_PATH=$(LIBDIR1):$$LD_LIBRARY_PATH ./textnormalizer $(RESULTDIR) $(LANG)

runandselfclean: run clean
	rm -f textnormalizer.pyx Makefile WikipediaMarkupCleaner.jar
	rm -fr nltk-master/