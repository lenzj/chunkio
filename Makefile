.POSIX:

PNAME = chunkio

RTEMPLATE ?= ../repo-template

all: check

.DEFAULT_GOAL := all

.PHONY: all

#---Golang Library Section---

GO ?= go
GOFLAGS ?=
GOSRC != find . -name '*.go'

check: $(GOSRC)
	$(GO) test $(GOFLAGS)

.PHONY: check

#---Generate Main Documents---

regenDocMain:
	pgot -i ":$(RTEMPLATE)" -o README.md template/README.md.got
	pgot -i ":$(RTEMPLATE)" -o LICENSE $(RTEMPLATE)/LICENSE.src/BSD-2-clause.got

.PHONY: regenDocMain

#---Generate Makefile---

regenMakefile:
	pgot -i ":$(RTEMPLATE)" -o Makefile template/Makefile.got

.PHONY: regenMakefile

#---Lint Helper Target---

lint:
	@find . -path ./.git -prune -or \
		-type f -and -not -name 'Makefile' \
		-exec grep -Hn '<no value>' '{}' ';'

.PHONY: lint

#---TODO Helper Target---

todo:
	@find . -path ./.git -prune -or \
		-type f -and -not -name 'Makefile' \
		-exec grep -Hn TODO '{}' ';'

.PHONY: todo

# vim:set noet tw=80:
