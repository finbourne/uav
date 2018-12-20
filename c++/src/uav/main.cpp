#include <iostream>
#include <fstream>
#include "string.hpp"
#include "failure.hpp"
#include "uav.hpp"

static void
write_output
(const std::string &path, const std::string &output)
{
  if ("" == path)
    throw new uav::failure("Invalid (empty) output path supplied");

  if ("-" == path)
  {
    std::cout << output << "\n";
    return;
  }

  auto file = std::ofstream(path, std::ofstream::out);
  file << output << "\n";
  file.close();
}

int
main
(int argc, char **argv)
{
  uav::arguments arguments;
  std::string output_path = "pipeline.json";

  //read args
  for (int i=1; i<argc; i++)
  {
    const auto arg = std::string(argv[i]);
    if (arg == "-h" || arg == "--help")
    {
      std::cout << "help\n";
      return 0;
    }

    if (arg == "-c" || arg == "--credentials")
    {
      const auto credentials_path = std::string(argv[i+1]);
      if (credentials_path == "")
      {
        std::cerr << "Error: No credentials supplied.\n";
        return 1;
      }

      if (arguments.credentials_path != "")
      {
        std::cerr << "Error: Only one credentials set is supported.";
        return 1;
      }

      arguments.credentials_path = credentials_path;
      i += 1;
      continue;
    }

    if (arg == "-d" || arg == "--define")
    {
      const auto substitution = std::string(argv[i+1]);
      if (substitution == "")
      {
        std::cerr << "Error: No substitution defined.\n";
        return 1;
      }

      const auto kvp = parse_substitution(substitution);
      if (kvp.first == "")
      {
        std::cerr << "Error: Malformed substitution definition.\n";
        return 1;
      }

      arguments.substitutions[kvp.first] = kvp.second;
      i += 1;
      continue;
    }

    if (arg == "-o" || arg == "--output")
    {
      const auto output = std::string(argv[i+1]);
      if (output == "")
      {
        std::cerr << "Error: No output defined.\n";
        return 1;
      }

      output_path = output;
      i += 1;
      continue;
    }

    arguments.input_paths.push_back(argv[i]);
  }

  if (0 == arguments.input_paths.size())
  {
    std::cerr << "Error: No pipeline definitions supplied.";
    return 1;
  }

  try {
    write_output(
      output_path,
      generate_pipeline(arguments)
    );

    return 0;
  }
  catch(const std::exception &e)
  {
    std::cerr
      << "Error: Unable to generate pipelines:\n"
      << e.what()
      << "\n";

    return 1;
  }
}
