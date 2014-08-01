{{if .Del }}
This post deleted by owner
{{else}}
{{if and (gt .Tid 1) (ne .Tid .LoginId)}}
This post is not for you, sorry
{{end}}
{{end}}
{{ if and (gt .Cle 1) .Sub }}
{{.Who}} at {{.When}}<div class="subject">{{if .Afile }}<a target="post"
href="/phh/atta.phh?cle={{.Cle}}"><img border=0 src="/static/pclip.gif"
 alt="clip" title="{{.Afile}}"/></a> {{end}}Subject:
{{.Sub}}
{{ if gt .LoginId 0 }}
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
<a href="/phh/post.phh?cle={{.Cle}}&cool=1">Cool</a>
{{if .IsPic }}</br><img border=0 src="/phh/atta.phh?cle={{.Cle}}"
 alt="{{.Afile}}"/>{{end}}
{{end}}</div>
<div class="body">{{.Body}}</div><br>
{{end}}
{{if and (eq .LoginId .Uid) .LoginId }}
<form method="post" action="/phh/post.phh">
<input type="hidden" name="delcle" value="{{ .Cle }}">
<input type="submit" value="{{ if .Del }}Undelete{{else}}Delete{{end}}"></form>
{{end}}
{{ if gt .LoginId 0 }}
<form method="post" action="/phh/post.phh" enctype="multipart/form-data">
<input type=hidden name="parent" value="{{ .Cle }}">
Login as [{{ .LoginName }}]: <a href="/phh/reg.phh?logout=1">logout</a><br>
Subject: <input type=text size=48 name="subject" value="">
<input type=submit value="Post"><br>
Only to: <input type=text size=24 name="onlyto" value="{{.Tow}}">
(not more than one ID)<br>
<textarea name="body" cols=53 rows=6></textarea><br>
<input type=submit value="Post"><br>
Attachment: <input type=file name="attachment"><br>
</form>
{{else}}
<form method="post" action="/phh/post.phh" enctype="multipart/form-data">
Name: <input type=text name="username">
Password: <input type=password name="password">
<input type=submit value="Login"><br>
</form>
{{end}}
