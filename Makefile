BIN_DIR := bin
CMD_DIRS := $(patsubst %/main.go,%,$(wildcard */main.go))
DMN_SRCS := $(foreach d,$(CMD_DIRS),$(wildcard $(d)/*.dmn))

BINARIES := $(addprefix $(BIN_DIR)/,$(CMD_DIRS))
DMNS := $(foreach f,$(DMN_SRCS),$(BIN_DIR)/$(notdir $(f)))

LINUX_AMD64_DIR := $(BIN_DIR)/linux_amd64
LINUX_AMD64_BINARIES := $(addprefix $(LINUX_AMD64_DIR)/,$(CMD_DIRS))
LINUX_AMD64_DMNS := $(foreach f,$(DMN_SRCS),$(LINUX_AMD64_DIR)/$(notdir $(f)))

.PHONY: all build build-linux-amd64 clean

all: build build-linux-amd64

build: $(BINARIES) $(DMNS)

build-linux-amd64: $(LINUX_AMD64_BINARIES) $(LINUX_AMD64_DMNS)

$(BIN_DIR) $(LINUX_AMD64_DIR):
	mkdir -p $@

define BUILD_BIN
$(BIN_DIR)/$(1): $(1)/main.go $(wildcard $(1)/*.dmn) | $(BIN_DIR)
	go build -o $$@ ./$(1)
endef
$(foreach d,$(CMD_DIRS),$(eval $(call BUILD_BIN,$(d))))

define BUILD_BIN_LINUX_AMD64
$(LINUX_AMD64_DIR)/$(1): $(1)/main.go $(wildcard $(1)/*.dmn) | $(LINUX_AMD64_DIR)
	GOOS=linux GOARCH=amd64 go build -o $$@ ./$(1)
endef
$(foreach d,$(CMD_DIRS),$(eval $(call BUILD_BIN_LINUX_AMD64,$(d))))

define COPY_DMN
$(BIN_DIR)/$(notdir $(1)): $(1) | $(BIN_DIR)
	cp $(1) $$@
endef
$(foreach f,$(DMN_SRCS),$(eval $(call COPY_DMN,$(f))))

define COPY_DMN_LINUX_AMD64
$(LINUX_AMD64_DIR)/$(notdir $(1)): $(1) | $(LINUX_AMD64_DIR)
	cp $(1) $$@
endef
$(foreach f,$(DMN_SRCS),$(eval $(call COPY_DMN_LINUX_AMD64,$(f))))

clean:
	rm -rf $(BIN_DIR)
