cxx=g++
cxxflags=-std=c++17 -O0 -g -I./include -I./src/uav
ld=$(cxx)
ldflags=$(cxxflags) -lyaml-cpp
