#include <catch.hpp>
#include <string.hpp>

TEST_CASE(
  "uav::string::split",
  "[uav/string/split]"
)
{
  SECTION("no split parameters")
  {
    const auto string = uav::string("some string");
    const auto expected = uav::string::list { string };

    REQUIRE(string.split("") == expected );
    REQUIRE(string.split(std::set<char> { }) == expected);
  }

  SECTION("split empty string")
  {
    const auto string = uav::string("");
    const auto expected = uav::string::list { string };
    REQUIRE(string.split("") == expected );
    REQUIRE(string.split("a") == expected );

    REQUIRE(string.split(std::set<char> { }) == expected);
    REQUIRE(string.split(std::set<char> { 'a' }) == expected);
  }

  SECTION("simple test cases")
  {
    const auto string1 = uav::string("some string");
    const auto expected1 = uav::string::list {
      "some",
      "string"
    };

    REQUIRE(string1.split(" ") == expected1);
    REQUIRE(string1.split(std::set<char>{ ' ' }) == expected1);

    const auto expected2 = uav::string::list {
      "ome ", 
      "tring"
    };

    REQUIRE(string1.split("s") == expected2);
    REQUIRE(string1.split(std::set<char>{ 's' }) == expected2);

    const auto expected3 = uav::string::list {
      "some string"
    };
    REQUIRE(string1.split("q") == expected3);
    REQUIRE(string1.split(std::set<char>{ 'q' }) == expected3);

    const auto string3 = uav::string("some other string");
    const auto expected4 = uav::string::list {
      "some ",
      " string"
    };
    REQUIRE(string3.split("other") == expected4);
  }

  SECTION("edge cases")
  {
    const auto string = uav::string("some string");

    REQUIRE(string.split("some string") == uav::string::list { "" });
  }

  SECTION("larger cases")
  {
    const auto string = uav::string("larger:set:of:delimited::strings::");
    const auto expected = uav::string::list {
      "larger",
      "set",
      "of",
      "delimited",
      "strings",
    };

    REQUIRE(string.split(":") == expected);
    REQUIRE(string.split(std::set<char> { ':' }) == expected);
  }
}

TEST_CASE(
  "uav::string::replace",
  "[uav/string/replace]"
)
{
  SECTION("trivial replacements")
  {
    REQUIRE(uav::string("").replace("a", "b") == "");
    REQUIRE(uav::string("").replace("", "b") == "");
    REQUIRE(uav::string("a").replace("", "b") == "a");
    REQUIRE(uav::string("q").replace("a", "b") == "q");
    REQUIRE(uav::string("q").replace("a", "") == "q");
    REQUIRE(uav::string("a").replace("a", "") == "");
    REQUIRE(uav::string("a").replace("a", "b") == "b");
  }

  SECTION("larger replacements")
  {
    REQUIRE(uav::string("some string").replace("", "") == "some string");
    REQUIRE(uav::string("some string").replace("a", "") == "some string");
    REQUIRE(uav::string("some string").replace("s", "b") == "bome btring");
    REQUIRE(uav::string("some string").replace("some", "other") == "other string");
  }

  SECTION("large replacement")
  {
    const auto string = uav::string(R"(
some ((string)) that needs replacing with another ((string))
read from a set of ((strings)) somewhere else.
    )");

    const auto expected = uav::string(R"(
some duck that needs replacing with another duck
read from a set of ((strings)) somewhere else.
    )");

    REQUIRE(string.replace("((string))", "duck") == expected);
  }
}

TEST_CASE(
  "uav::string::count",
  "[uav/string/count]",
)
{
  SECTION("trivial counts")
  {
    REQUIRE(uav::string("").count("") == 0);
    REQUIRE(uav::string("a").count("") == 0);
    REQUIRE(uav::string("").count("a") == 0);
    REQUIRE(uav::string("").count(std::set<char>{ }) == 0);
    REQUIRE(uav::string("").count(std::set<char>{ 'a' }) == 0);
    REQUIRE(uav::string("a").count(std::set<char>{ }) == 0);
  }

  SECTION("pattern counts")
  {
    REQUIRE(uav::string("some string").count("s") == 2);
    REQUIRE(uav::string("some string").count("so") == 1);
    REQUIRE(uav::string("some string").count(" s") == 1);
    REQUIRE(uav::string("some string string").count("string") == 2);
  }

  SECTION("set counts")
  {
    REQUIRE(uav::string("some string").count(std::set<char>{ 's', 't' }) == 3);
    REQUIRE(uav::string("some string").count(std::set<char>{ 's', 'q' }) == 2);
    REQUIRE(uav::string("some string").count(std::set<char>{ 'q' }) == 0);
  }
}
