# comparinator

Simple visual regression testing for websites with reasonable defaults.

If you need something more advanced, wraith is probably what you want. This tool will only crawl the base path and any pages on the alpha site that are directly linked from the base path (determined by domain).

## Usage

`comparinator -help` shows all available flags.

For the most common testing scenario:

`comparinator -alpha-base-url="https://www.yoursite.com" -beta-base-url="https://dev.yoursite.com"`

You'll need to have webdriver running at `localhost:4444` or specify the URL to your webdriver with the `-webdriver-url` flag.

Once it's done running, you'll have an output directory with all of the screenshots + diffs + a JSON file with a bunch of info in it about which screenshots belong to which url, how similar they are, and the overall similarity across all screenshots. Someday I'll build a little web UI around this.

## Building

`make`

## Releasing

Push a tag to the repo and Github CI will take over, build a release, and upload
the artifacts to the release page.
