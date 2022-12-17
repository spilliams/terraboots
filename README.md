# terraboots

My attempt at a Terraform build orchestrator, for large platform projects with
hundreds of root modules.

This repository contains both the source code for the tool `terraboots`, as well
as a sample monorepo managed by that tool.

## Example Monorepo

The main configuration file for the monorepo is `terraboots.hcl`. This defines
where our root configurations live.

Each of these roots has its own `terraboots.hcl` which contains configuration
details about that root.

The main concept to terraboots is that it expects you to manage your platform
with many small terraform root configurations. You do this through `scope`s that
you define in the top level hcl, and selectively apply to each of your roots.

This allows you to maintain a few root "templates" that each could be planned
and applied dozens or hundreds of times depending on the permutations of your
scopes.

## Terraboots CLI

See `src/README.md`.
