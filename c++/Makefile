include make/opts.make

#list all folders in src
uav_binaries=$(addprefix bin/,$(shell ls src))

# list all files in tests that aren't uav.tests.base.cpp
test_binaries=$(subst tests/,tests/bin/,$(filter-out %base,$(subst .cpp,,$(shell ls tests/*tests.cpp))))

all: $(uav_binaries)

bin/uav: $(subst .cpp,.o,$(subst src/,obj/,$(shell find src/uav -name '*.cpp')))
	@echo "LD  $@"; mkdir -p $(dir $@); $(ld) -o $@ $^ $(ldflags)

obj/%.o: src/%.cpp
	@echo "CXX $<"; mkdir -p $(dir $@); $(cxx) $(cxxflags) -c -o $@ $^

tests: $(test_binaries)
	@$(foreach test,$^,echo "Running $(test)"; $(test);)

tests/bin/%: tests/obj/%.o tests/obj/uav.tests.base.o
	@echo "LD  $@"; mkdir -p $(dir $@); $(ld) -o $@ $^ $(ldflags)

tests/obj/uav.atc.%.tests.o: tests/uav.atc.%.tests.cpp include/atc/%.hpp
	@echo "CXX $<"; mkdir -p $(dir $@); $(cxx) $(cxxflags) -c -o $@ $<

tests/obj/uav.%.tests.o: tests/uav.%.tests.cpp include/%.hpp
	@echo "CXX $<"; mkdir -p $(dir $@); $(cxx) $(cxxflags) -c -o $@ $<

tests/obj/%.o: tests/%.cpp
	@echo "CXX $<"; mkdir -p $(dir $@); $(cxx) $(cxxflags) -c -o $@ $<

clean:
	@echo "CLEAN"
	@rm -fr tests/obj
	@rm -fr tests/bin
	@rm -fr obj
	@rm -fr bin

.PHONY: all tests clean .test-objs .bin-objs

#this pile of garbage is to encourage make not to delete object files
.test-objs: $(subst tests/,tests/obj/,$(subst .cpp,.o,$(shell ls tests/*.cpp)))
	;
.bin-objs: $(subst obj/,src/,$(subst .cpp,.o,$(shell find src -name '*.cpp')))
	;
