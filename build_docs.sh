#!/bin/bash

parent_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd) || exit

cd "$parent_dir"/wiki || exit

source .venv/bin/activate

mkdocs build
