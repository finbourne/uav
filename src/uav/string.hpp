#pragma once
#include <string>
#include <vector>
#include <set>
#include <stdint.h>

namespace uav {
  class string: public std::string {
    public:
    typedef std::vector<uav::string> list;

    template<typename ...tArgs>
    string
    (tArgs &&...args):
    std::string(std::forward<tArgs>(args)...)
    { }

    std::vector<uav::string>
    split
    (const std::set<char> &delimiters)
    const noexcept
    {
      if (0 == delimiters.size())
        return { *this };

      if (0 == length())
        return { *this };

      const std::string charset(delimiters.begin(), delimiters.end());

      std::vector<uav::string> result;
      size_t pos = 0;

      while (pos < length()) {
        const size_t match = find_first_of(charset, pos);
        if (match != pos)
          result.push_back(substr(pos, match - pos));

        if (match == std::string::npos)
          break;

        pos = match + 1;
      }

      if (0 == result.size())
        result.push_back("");

      return result;
    }

    std::vector<uav::string>
    split
    (const std::string &pattern)
    const noexcept
    { 
      if (0 == pattern.length())
        return { *this };

      std::vector<uav::string> result;
      size_t pos = 0;
      while (pos < length()) {
        const size_t match = find(pattern, pos);
        if (match != pos)
          result.push_back(substr(pos, match - pos));

        if (match == std::string::npos)
          break;

        pos = match + pattern.length();
      }

      if (0 == result.size())
        result.push_back("");

      return result;
    }

    uav::string
    replace
    (const std::string &pattern, const std::string &replacement)
    const noexcept
    {
      if ("" == pattern)
        return *this;

      std::string result = "";
      size_t pos = 0;
      
      while (pos < length()) {
        const size_t match = find(pattern, pos);
        result += substr(pos, match - pos);

        if (match == std::string::npos)
          break;

        result += replacement;

        pos = match + pattern.length();
      }

      return result;
    }

    uint32_t
    count
    (const std::set<char> &charset)
    const noexcept
    {
      if (0 == charset.size())
        return 0;
      
      size_t result = 0;
      for (const char c: *this)
        result += charset.count(c);

      return result;
    }

    uint32_t
    count
    (const std::string &pattern)
    const noexcept
    {
      if ("" == pattern)
        return 0;

      uint32_t result = 0;
      for (size_t p=find(pattern); std::string::npos != p; p=find(pattern, p + 1))
        result += 1;

      return result;
    }

    bool
    starts_with
    (const std::string &str)
    const noexcept
    {
      if (str.length() > length())
        return false;

      return (substr(0, str.length()) == str);
    }

    bool
    ends_with
    (const std::string &str)
    const noexcept
    {
      if (str.length() > length())
        return false;

      return (substr(length() - str.length()) == str);      
    }
  };
}

