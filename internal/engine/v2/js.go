package engine

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func injectOpenActionJS(ctx *model.Context, start, end time.Time, experiredText, unsupportedText string) {
	escapedExpiredText := escapeJSString(experiredText)

	js := fmt.Sprintf(`(function(){
  try{
    if(this.__ocg_js_executed) return; this.__ocg_js_executed = true;
    var start = new Date("%s");
    var end = new Date("%s");
    var now = new Date();
    var inRange = (now >= start && now <= end);
    try {
      if(!inRange){
        app.alert({
            cMsg: "%s",
          });	
        this.closeDoc(true);
        return;  
      }
     if (typeof this.getOCGs === "function") {
        var ocgs = this.getOCGs();
        for (var i = 0; i < ocgs.length; i++) {
          var o = ocgs[i];
          o.state = false;
        }
      }
    } catch (e) { app.alert(e); }
    } catch (e) { }
})();`, start.Format(time.RFC3339), end.Format(time.RFC3339), escapedExpiredText)

	// 假设 encodeJSUTF16BE 返回 []byte（UTF‑16BE 带 BOM）
	utf16Bytes := encodeJSUTF16BE(js)

	// 编成十六进制放进 HexLiteral
	hexStr := hex.EncodeToString(utf16Bytes)

	iref, err := ctx.IndRefForNewObject(types.Dict{
		"S":  types.Name("JavaScript"),
		"JS": types.HexLiteral(hexStr),
	})
	if err != nil {
		fmt.Printf("injectOpenActionJS: %v\n", err)
	}
	ctx.RootDict["OpenAction"] = *iref
}

// ...existing code...
