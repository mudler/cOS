#
# cOS-toolkit Makefile
#
#

#----------------------- global variables -----------------------
#
# Path to luet binary
#
export LUET?=$(shell which luet 2> /dev/null)
ifeq ("$(LUET)","")
LUET="/usr/bin/luet"
endif

#
# Path to jq binary
#
export JQ?=$(shell which jq 2> /dev/null)
ifeq ("$(JQ)","")
JQ="/usr/bin/jq"
endif

#
# Path to yq binary
#
export YQ?=$(shell which yq 2> /dev/null)
ifeq ("$(YQ)","")
YQ="/usr/bin/yq"
endif

#
# Path to luet-makeiso binary
#
export MAKEISO?=$(shell which luet-makeiso 2> /dev/null)
ifeq ("$(MAKEISO)","")
MAKEISO="/usr/bin/luet-makeiso"
endif

#
# Output for "make publish-repo" and base for "make iso"
#
FINAL_REPO?=raccos/releases-$(FLAVOR)

#
# Location of package tree
#
TREE?=$(ROOT_DIR)/packages

#
# Directory of Makefile
#
export ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

#
# OS flavor to build
#
FLAVOR?=opensuse

#
# folder for build artefacts
#
DESTINATION?=$(ROOT_DIR)/build

#
# yaml specification of build targets
#
export ISO_SPEC?=$(ROOT_DIR)/iso/cOS.yaml

#----------------------- end global variables -----------------------


#----------------------- default target -----------------------

all: deps build

#----------------------- includes -----------------------

include make/Makefile.build
include make/Makefile.iso
include make/Makefile.run
include make/Makefile.test

#----------------------- targets -----------------------

deps: $(LUET) $(YQ) $(JQ) $(MAKEISO)

$(LUET):
ifneq ($(shell id -u), 0)
	@echo "'$@' is missing and you must be root to install it."
	@exit 1
else
	curl https://get.mocaccino.org/luet/get_luet_root.sh |  sh
	luet install -y repository/mocaccino-extra-stable
endif

$(YQ):
ifneq ($(shell id -u), 0)
	@echo "'$@' is missing and you must be root to install it."
	@exit 1
else
	$(LUET) install -y utils/yq
endif

$(JQ):
ifneq ($(shell id -u), 0)
	@echo "'$@' is missing and you must be root to install it."
	@exit 1
else
	$(LUET) install -y utils/jq
endif

$(MAKEISO):
ifneq ($(shell id -u), 0)
	@echo "'$@' is missing and you must be root to install it."
	@exit 1
else
	$(LUET) install -y extension/makeiso
endif

clean: clean_build clean_iso clean_run clean_test
	rm -rf $(ROOT_DIR)/*.sha256
