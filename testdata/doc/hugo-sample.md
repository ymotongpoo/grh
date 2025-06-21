# Hugo ショートコードサンプル

このドキュメントはjqueryとcookieを含むHugoショートコードのテストです。

## コードハイライト

{{< highlight javascript >}}
// jquery example with cookie
$('#cookie-banner').show();
var jquery = 'test';
{{< /highlight >}}

## 注意書き

{{% note %}}
このjqueryのcookieサンプルは重要です。
{{% /note %}}

## 図表

{{< figure src="jquery-cookie-diagram.jpg" alt="jQuery Cookie の図" >}}

## インラインコード

通常のテキストでjqueryとcookieを使用する場合は置換されます。

```javascript
// このコードブロック内のjqueryとcookieも置換されます
console.log('jquery cookie example');
```

## 複雑なショートコード

{{< highlight go "linenos=table,hl_lines=2 3" >}}
package main
import "fmt" // jquery
func main() { // cookie
    fmt.Println("Hello World")
}
{{< /highlight >}}

{{% expand "詳細情報" %}}
ここでもjqueryとcookieが置換されるはずです。
{{% /expand %}}
