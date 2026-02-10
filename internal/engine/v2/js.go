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
    var alertMsg = function(msg){
        if (typeof app !== "undefined" && app && typeof app.alert === "function") {
            app.alert({ cMsg: msg });
        }
  	};
    // debugger;
    // alertMsg("有效期PDF");
    var start = new Date("%s");
    var end = new Date("%s");
    var now = new Date();
    var inRange = (now >= start && now <= end);
	
    var myOCGs = function(){
        if (typeof getOCGs === "function") {
            return getOCGs();
        }
        if (typeof this.getOCGs === "function") {
            return this.getOCGs();
        }
        return null;
    };
    // 关闭所有 OCG
    var closeAllOCGs = function(){
        var ocgs = myOCGs();
        if (ocgs && ocgs.length) {
            for (var i = 0; i < ocgs.length; i++) {
              if (ocgs[i]) {
                // 关闭不是水印的 OCG
                if (!ocgs[i].name || ocgs[i].name.indexOf("Watermark") !== 0) {
                  ocgs[i].state = false;
                }
              }
            }
        }
    };
    // 关闭text_*OCG
    var closeTextOCGs = function(){
        var ocgs = myOCGs();
        if (ocgs && ocgs.length) {
            for (var i = 0; i < ocgs.length; i++) {
              if (ocgs[i] && ocgs[i].name && ocgs[i].name.indexOf("text_") === 0) {
                ocgs[i].state = false;
              }
              //关闭 expired_mask_* OCG
              if (ocgs[i] && ocgs[i].name && ocgs[i].name.indexOf("expired_mask_") === 0) {
                ocgs[i].state = false;
              }
            }
        }
    };
      if(!inRange){
        // 过期关闭text提示
        closeTextOCGs();
        if("%s" !== ""){
          alertMsg("%s");
        }
		if (this && typeof this.closeDoc === "function") {
        	this.closeDoc(true);
			 return;  
      	}
		if (typeof closeDoc === "function"){
			closeDoc(true);
			return;
		}
        return;
      }
      // zh: 在有效期内，关闭所有 OCG
      closeAllOCGs();
      return;
    } catch (e) {}
})();`, start.Format(time.RFC3339), end.Format(time.RFC3339), escapedExpiredText, escapedExpiredText)

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
