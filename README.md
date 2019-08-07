# Negapedia Conflict Analyzer
[![Go Report Card](https://goreportcard.com/badge/github.com/MarcoChilese/NegapediaConflictualWords_GO)](https://goreportcard.com/report/github.com/MarcoChilese/NegapediaConflictualWords)

## Usage
#### Building docker image
``docker build -t <image_name> .``<br>
from the root of repository directory.

#### Running docker image
``docker run -d -v <path_on_fs_where_to_save_results>:<container_results_path> <image_name>``<br>

#### Dockerfile flags
- `-l`: wiki language;
- `-d`: container result dir;
- `-s`: revert starting date to consider;
- `-e`: revert ending date to consider;
- `-specialList`: special page list to consider;
- `-r`: number of revert to consider;
- `-t`: number of top words per page to consider;
