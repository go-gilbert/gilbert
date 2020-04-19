PKG := github.com/go-gilbert/gilbert/v2/cmd/gb

define gb
	@go run $(PKG) $(1) $(2)
endef


.PHONY: run
run:
	$(call gb,run,ls)

.PHONY: ls
ls:
	$(call gb,ls)

.PHONY: gb
gb:
	$(call gb,run,$(name))