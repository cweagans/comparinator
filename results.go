package main

import (
	"bytes"
	"text/template"
)

const resultsHTMLTemplate = `
<!doctype html>
<html>
	<head>
		<title>{{ .Title }} | Comparinator</title>
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/semantic-ui@2.4.2/dist/semantic.min.css" integrity="sha256-UXesixbeLkB/UYxVTzuj/gg3+LMzgwAmg3zD+C4ZASQ=" crossorigin="anonymous">
		<style>
			.horizontal.divider.header {
				margin-top: 40px;
				margin-bottom: 40px;
			}
			img {
				max-width: 100%;
			}
		</style>
	</head>
	<body>
		<div class="ui very padded segment container">
			<h1 class="ui header">{{ .Title }}</h1>
			
			<h4 id="info" class="ui horizontal divider header"><i class="blue info icon"></i>Information</h4>
			<table class="ui definition table">
				<tbody>
				<tr>
					<td class="three wide column">Start time</td>
					<td>{{ .CompareStartTimeFormatted }}</td>
				</tr>
				<tr>
					<td>End time</td>
					<td>{{ .CompareEndTimeFormatted }}</td>
				</tr>
				<tr>
					<td>Alpha URL</td>
					<td><a href="{{ .AlphaBaseURL }}">{{ .AlphaBaseURL }}</a></td>
				</tr>
				<tr>
					<td>Beta URL</td>
					<td><a href="{{ .BetaBaseURL }}">{{ .BetaBaseURL }}</a></td>
				</tr>
				<tr>
					<td>Base path</td>
					<td>{{ .TestPath }}</td>
				</tr>
				<tr>
					<td>Pages compared</td>
					<td>{{ len .Links }}</td>
				</tr>
				<tr>
					<td>Overall similarity</td>
					<td><strong>{{ .OverallSimilarity }}%</strong></td>
				</tr>
				</tbody>
			</table>
			
			<h4 id="differing" class="ui horizontal divider header"><i class="red search icon"></i>Differing pages</h4>
			<div class="ui stackable two column grid container">

			{{ range $path, $link := .Links }}
		
				{{ if lt $link.Similarity 100.0 }}

				<div class="column">
					<div class="ui fluid card">
						<div class="image">
							<a href="{{ $link.DiffFile }}"><img src="{{ $link.DiffFile }}"></a>
						</div>
						<div class="content">
							<div class="description">
								<table class="ui definition table">
									<tbody>
										<tr>
											<td>Path</td>
											<td>{{ $path }}</td>
										</tr>
										<tr>
											<td>Similarity</td>
											<td><strong>{{ $link.Similarity }}%</strong></td>
										</tr>
										<tr>
											<td class="three wide column">Alpha link</td>
											<td><a href="{{ .AlphaBaseURL }}{{ $path }}">{{ .AlphaBaseURL }}{{ $path }}</a></td>
										</tr>
										<tr>
											<td class="three wide column">Beta link</td>
											<td><a href="{{ .BetaBaseURL }}{{ $path }}">{{ .BetaBaseURL }}{{ $path }}</a></td>
										</tr>
									</tbody>
								</table>
							</div>
						</div>
					</div>
				</div>

				{{ end }}

			{{ end }}
				
			</div>
			
			<h4 id="other" class="ui horizontal divider header"><i class="green check icon"></i>Other pages</h4>
			<div class="ui stackable three column doubling grid container">

			{{ range $path, $link := .Links }}
				{{ if eq $link.Similarity 100.0 }}

				<div class="column">
					<div class="ui fluid card">
						<div class="image">
							<a href="{{ $link.AlphaScreenshotFile }}"><img src="{{ $link.AlphaScreenshotFile }}"></a>
						</div>
						<div class="content">
							<div class="description">
								<table class="ui definition table">
									<tbody>
										<tr>
											<td>Path</td>
											<td>{{ $path }}</td>
										</tr>
										<tr>
											<td>Similarity</td>
											<td><strong>{{ $link.Similarity }}%</strong></td>
										</tr>
									</tbody>
								</table>
							</div>
						</div>
					</div>
				</div>

				{{ end }}
			{{ end }}

			</div>
		</div>
	</body>
</html>

`

func getResultsHTML(result Result) ([]byte, error) {

	t, err := template.New("results").Parse(resultsHTMLTemplate)
	if err != nil {
		return make([]byte, 0), err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, result)
	if err != nil {
		return make([]byte, 0), err
	}

	return buf.Bytes(), nil
}
