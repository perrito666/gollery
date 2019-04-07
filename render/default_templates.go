package render

var singleImageDefaultTemplate = []byte(`<html>

<head>
    <title>{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</title>
    <description>{{.Description}}</description>
</head>
<nav class="breadcrumb">
    <p>
        {{range .ParentTree}}
        <a href="{{.Path}}">{{.FolderName}}</a>/
        {{end}}
    </p>
</nav>
<nav class="tree">
    {{/* Device a logic for current so we can add a class */}}
    {{range .Siblings}}
    <a href="{{.FileName}}" class="sibling_picture">{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</a>
    {{end}}
    {{range .Children}}
    <a href="{{.FolderName}}" class="sibling_album">{{if .Title}} {{.Title}} {{else}} {{.FolderName}} {{end}}</a>
    {{end}}
</nav>
<nav class="relative">
    <a href="{{.First}}" class="first_picture">{{if .First.Title}} {{.First.Title}} {{else}} {{.First.FileName}}
        {{end}}</a>
    <a href="{{.Previous}}" class="previous_picture">{{if .Previous.Title}} {{.Previous.Title}} {{else}}
        {{.Previous.FileName}} {{end}}</a>
    <a href="{{.Next}}" class="next_picture">{{if .Next.Title}} {{.Next.Title}} {{else}} {{.Next.FileName}} {{end}}</a>
    <a href="{{.Last}}" class="last_picture">{{if .Last.Title}} {{.Last.Title}} {{else}} {{.Last.FileName}} {{end}}</a>
</nav>
<header>
    <h1>{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</h1>
</header>
<main>
    <img src="{{.FileName}}" alt="{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}">
    <p>{{.Description}}</p>
</main>
<footer></footer>

</html>
`)

var groupDefaultTemplate = []byte(`<html>

<head>
    <title>{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</title>
    <description>{{.Description}}</description>
</head>
<nav class="breadcrumb">
    <p>
        {{range .ParentTree}}
        <a href="{{.Path}}">{{.FolderName}}</a>/
        {{end}}
    </p>
</nav>
<nav class="tree">
    {{/* Device a logic for current so we can add a class */}}
    {{range .Siblings}}
    <a href="{{.FileName}}" class="sibling_picture">{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</a>
    {{end}}
    {{range .Children}}
    <a href="{{.FolderName}}" class="sibling_album">{{if .Title}} {{.Title}} {{else}} {{.FolderName}} {{end}}</a>
    {{end}}
</nav>
<nav class="relative">
    <a href="{{.First}}" class="first_picture">{{if .First.Title}} {{.First.Title}} {{else}} {{.First.FileName}}
        {{end}}</a>
    <a href="{{.Previous}}" class="previous_picture">{{if .Previous.Title}} {{.Previous.Title}} {{else}}
        {{.Previous.FileName}} {{end}}</a>
    <a href="{{.Next}}" class="next_picture">{{if .Next.Title}} {{.Next.Title}} {{else}} {{.Next.FileName}} {{end}}</a>
    <a href="{{.Last}}" class="last_picture">{{if .Last.Title}} {{.Last.Title}} {{else}} {{.Last.FileName}} {{end}}</a>
</nav>
<header>
    <h1>{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}</h1>
</header>
<main>
    {{range .Images}}
    <section class="thumb">
        <a href="{{.FileName}}">
            {{/* thumbs are obtaining by GETing FileName___widthxheight assuming said size is permitted if it is not you just get the default one */}}
            <img src="{{.FileName}}___640x480" alt="{{if .Title}} {{.Title}} {{else}} {{.FileName}} {{end}}">
            <p>{{.Description}}{{/* perhaps this could be truncated */}}</p>
        </a>
    </section>
    {{end}}
</main>
<footer></footer>

</html>`)
