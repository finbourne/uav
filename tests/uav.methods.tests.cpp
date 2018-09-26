#include <catch.hpp>

//including .cpp directly
#include <uav.cpp>

TEST_CASE(
  "uav::methods::add_all_except",
  "[uav/uav/add_all_except]"
)
{
  const json initial = {
    { "key1", "initial1" },
  };

  const json remaining = {
    { "key1", "remaining1" },
    { "key2", "remaining2" },
    { "key3", "remaining3" },
  };

  SECTION("no except clause")
  {
    const json expected = {
      { "key1", "initial1" },
      { "key2", "remaining2" },
      { "key3", "remaining3" },
    };
    REQUIRE(uav::add_all_except(initial, remaining) == expected);
  }

  SECTION("with except clause, overlapping")
  {
    const json expected = {
      { "key1", "initial1" },
      { "key3", "remaining3" },
    };

    REQUIRE(uav::add_all_except(initial, remaining, { "key2" }) == expected);
  }
}

TEST_CASE(
  "uav::methods::get_directory",
  "[uav/uav/get_directory]"
)
{
  SECTION("no directory")
  {
    REQUIRE(uav::get_directory("") == "");
    REQUIRE(uav::get_directory("some-file.txt") == "");
  }

  SECTION("directories")
  {
    REQUIRE(uav::get_directory("a/some-file.txt") == "a/");
    REQUIRE(uav::get_directory("a/b/some-file.txt") == "a/b/");
  }
}

TEST_CASE(
  "uav::methods::get_path_relative_to",
  "[uav/uav/get_path_relative_to]"
)
{
  SECTION("trivial paths")
  {
    REQUIRE(uav::get_path_relative_to("some-file.txt") == "some-file.txt");
    REQUIRE(uav::get_path_relative_to("a/some-file.txt") == "a/some-file.txt");
  }

  SECTION("simple relative paths")
  {
    const uav::string::list lineage = {
      "a/file.txt",
    };

    REQUIRE(uav::get_path_relative_to("some-file.txt", lineage) == "a/some-file.txt");
    REQUIRE(uav::get_path_relative_to("b/some-file.txt", lineage) == "a/b/some-file.txt");
  }

  SECTION("deep relative paths")
  {
    const uav::string::list lineage = {
      "a/file.txt",
      "b/file.txt",
      "c/file.txt",
      "../file.txt",
    };

    REQUIRE(uav::get_path_relative_to("some-file.txt", lineage) == "a/b/c/../some-file.txt");
    REQUIRE(uav::get_path_relative_to("c/some-file.txt", lineage) == "a/b/c/../c/some-file.txt");
  }
}

TEST_CASE(
  "uav::methods::parse_substitution",
  "[uav/uav/parse_substitution]"
)
{
  typedef std::pair<std::string, std::string> string_pair;
  const string_pair empty = { "", "" };

  SECTION("empty case")
  {
    REQUIRE(parse_substitution("") == empty);
  }

  SECTION("valid cases")
  {
    REQUIRE(parse_substitution("a=b") == string_pair("a", "b"));
    REQUIRE(parse_substitution("a= b") == string_pair("a", " b"));
    REQUIRE(parse_substitution("a=") == string_pair("a", ""));
    REQUIRE(parse_substitution("a=") == string_pair("a", ""));
  }

  SECTION("invalid cases")
  {
    REQUIRE(parse_substitution("=b") == empty);
    REQUIRE(parse_substitution("asd") == empty);
  }
}
