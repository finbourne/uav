#pragma once

#include <string>
#include <exception>

namespace uav {
  class failure: public std::exception {
    std::string mDetail;

    public:
    failure
    (const std::string &detail):
    mDetail(detail)
    { }

    static uav::failure
    with
    (const std::string &detail)
    { return failure("  " + detail); }

    static uav::failure
    with
    (const std::exception &prev, const std::string &detail)
    { return failure("  " + detail + "\n" + std::string(prev.what())); }

    virtual const char *
    what(void)
    const noexcept
    { return mDetail.data(); }
  };
}
