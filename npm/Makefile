TARGET_DIRS = \
	./pkg/dataplane \
	./pkg/controlplane \
	./pkg/transport
all: generate

generate:
	for dir in $(TARGET_DIRS); do \
			make -C $$dir $@; \
	done

.PHONY: all generate