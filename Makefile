include Makefile.inc

.PHONY: subdirs $(DIRS)

$(DIRS):
	make -C $@
