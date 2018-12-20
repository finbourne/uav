#pragma once
#include "string.hpp"
#include <vector>
#include <map>

namespace uav {
  struct arguments {
    std::vector<std::string>            input_paths;
    std::map<std::string, std::string>  substitutions;
    std::string                         credentials_path;
  };

  std::string
  generate_pipeline
  (const uav::arguments &arguments);
}

static std::pair<std::string, std::string>
parse_substitution
(const uav::string &str)
{
  if (0 == str.count("="))
    return { "", "" };

  const auto split = str.split("=");
  if (split.size() == 0)
    return { "", "" };

  if (split.size() == 1)
  {
    if(split[0].length() == 0)
      return { "", "" };

    if (str[0] == '=')
      return { "", "" };

    return { split[0], "" };
  }

  if (0 == split[0].length())
    return { "", "" };

  std::string rhs = "";
  for (auto i = split.begin() + 1; i!=split.end(); i++)
    rhs += *i;

  return { split[0], rhs };
}
