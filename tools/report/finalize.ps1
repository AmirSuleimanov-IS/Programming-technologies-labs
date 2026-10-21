# Обновляет оглавление и поля в .docx через Word и сохраняет PDF рядом (для проверки)
param([string]$In, [string]$Pdf)
$word = New-Object -ComObject Word.Application
$word.Visible = $false
$word.DisplayAlerts = 0
$doc = $word.Documents.Open((Resolve-Path $In).Path)
foreach ($toc in $doc.TablesOfContents) { $toc.Update() }
$doc.Fields.Update() | Out-Null
$doc.Save()
if ($Pdf) { $doc.SaveAs2([ref]$Pdf, [ref]17) }
$doc.Close([ref]0)
$word.Quit()
