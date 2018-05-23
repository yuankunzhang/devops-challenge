# Github Stats

A simple Golang program that fetches properties for a given list of public Github repositories.

It reads the repository list from stdin. The list should be separated by new-lines. And each line should have this format: `$orgname/$repo`. Empty lines will be ignored and duplicates will be removed. Leading and trailing spaces in a line will be removed.

For example, this is a valid input:

```
octocat/hello-worId
kubernetes/charts

   torvalds/linux
kubernetes/charts
<EOF>
```

These properties are fetched for each repository:

- The name
- The clone URL
- The date of latest commit
- The name of latest author

**Note: It treats the latest commit as the latest commit in the default branch, which is not necessarily the master branch.**

## Dependency Management

It uses [dep](https://github.com/golang/dep) as dependency management tool.

## Building

### Building with Go

```shell
$ go build -o github-stats
```

You should now have the `github-stats` binary.

### Building with Docker

```shell
$ docker build -t github-stats:latest .
```

You should now have the `github-stats:latest` Docker image.

The multi-stages building technique is used in order to keep the image small (it is only 12.8MB).

## Environment

This program uses the [Github GraphQL API v4](https://developer.github.com/v4/).

Unlike the API v3, access token is a mandatory in order to use the API v4. A dummy access token with no permissions can do.

Before running the program, you need to [create a personal access token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/), and export it as the `GITHUB_ACCESS_TOKEN` environment variable.

```shell
$ export GITHUB_ACCESS_TOKEN=xxx
```

## Running

### Running the Binary

```shell
$ ./github-stats
octocat/hello-worId
kubernetes/charts
torvalds/linux
<EOF>

# OUTPUT:
Name,Clone URL,Date of Latest Commit,Name of Latest Author
hello-worId,https://github.com/octocat/hello-worId,2014-06-18T14:26:19-07:00,The Octocat
charts,https://github.com/kubernetes/charts,2018-05-22T08:48:54+01:00,Will Salt
linux,https://github.com/torvalds/linux,2018-05-22T09:00:00+10:00,Nicholas Piggin
```

Two options can be used along with the command.

```shell
$ ./github-stats -h
Usage of ./github-stats:
  -e	show errors
  -s	show summaries
```

For example:

```shell
$ ./github-stats -e -s
some-random-org
some-random-org/some-random-repo
octocat/hello-worId
<EOF>

# OUTPUT:
Name,Clone URL,Date of Latest Commit,Name of Latest Author
hello-worId,https://github.com/octocat/hello-worId,2014-06-18T14:26:19-07:00,The Octocat

Input Errors:
  <some-random-org> invalid input: should be in format of $orgname/$repo

Query Errors:
  <some-random-org/some-random-repo> query error: Could not resolve to a Repository with the name 'some-random-repo'.

Summaries:
  Total Unique Inputs (not including empty lines): 3
  Succeeded: 1
  Failed: 2
```

For testing purpose, I wrote a script named `gen_repo_list.sh`. It fetches hundreds of trending repositories from Github. To use this script (make sure you have [jq](https://github.com/stedolan/jq) installed):

```shell
$ ./gen_repo_list.sh | ./github-stats -e -s

# OUTPUT:
Name,Clone URL,Date of Latest Commit,Name of Latest Author
hub,https://github.com/github/hub,2018-05-18T14:50:06+02:00,Mislav Marohnić
dgraph,https://github.com/dgraph-io/dgraph,2018-05-18T09:19:26-07:00,Manish R Jain
# and a lot more...

Summaries:
  Total Unique Inputs (not including empty lines): 500
  Succeeded: 500
  Failed: 0
```

### Running with Docker

```shell
$ docker run --rm -it -e GITHUB_ACCESS_TOKEN=xxx github-stats:latest [-s, [-e]]
```

## Follow ups

Designing choices:

- I choose to use the GraphQL API instead of the Restful API, because it offers more flexibility to interact with Github. And it gives the client app a better performance (especially when the client app is large).
- The output order can be different from the input order, because of concurrency. We can sequentialize the API requests in order to keep the order, but the performance can be a big pain. Another approach to keep the order is to sort the result before printing it.
- The commit history is fetched from the default branch. Some repositories do not have a master branch.
- I choose to use `context.Background()` when sending requests. This is only a starting point, and needed to be improved.

Can be improved:

- Sorting the result to keep the output order the same as the input order.
- Fetching the commit history from a branch specified by the user.
- Using a custom context instead of `context.Background()`.
- Adding more fields to fetch, e.g. number of forks. This will be easy to achieve, thanks to the GraphQL API.

To be done:

- The rate limiting functionality is not implemented yet. According to the [doc](https://developer.github.com/v4/guides/resource-limitations/), this will be pretty straightforward.