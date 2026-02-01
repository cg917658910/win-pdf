package engine

import (
	"fmt"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func injectOpenActionJS(ctx *model.Context, start, end time.Time, experiredText, unsupportedText string) {

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
			cMsg: "%s"
			});	
		this.closeDoc(true);
        return;  
      }
     if (typeof this.getOCGs === "function") {
        var ocgs = this.getOCGs();
		app.alert("Found " + ocgs.length + " OCGs.");
        for (var i = 0; i < ocgs.length; i++) {
          var o = ocgs[i];
          o.state = false;
        }
      }
    } catch (e) { app.alert(%s + ' ' + e); }
    } catch (e) { }
})();`, start.Format(time.RFC3339), end.Format(time.RFC3339), experiredText, unsupportedText)

	iref, err := ctx.IndRefForNewObject(types.Dict{
		"S":  types.Name("JavaScript"),
		"JS": types.StringLiteral(encodeJSUTF16BE(js)),
	})
	if err != nil {
		fmt.Printf("injectOpenActionJS: %v\n", err)
	}

	/* ctx.RootDict["Names"] = types.Dict{
		"JavaScript": types.Dict{
			"Names": types.Array{
				types.StringLiteral("OpenActionJS"),
				*iref,
			},
		},
	} */

	// Set OpenAction to the JavaScript action indirect reference.
	ctx.RootDict["OpenAction"] = *iref
}
