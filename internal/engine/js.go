package engine

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func injectOpenActionJS(ctx *model.Context, start, end time.Time, experiredText, unsupportedText string) {

	// quote strings to be safely embedded into JS code
	qExpired := strconv.Quote(experiredText)
	qUnsupported := strconv.Quote(unsupportedText)

	js := fmt.Sprintf(`(function(){
  try{
    var start = new Date("%s");
    var end = new Date("%s");
    var now = new Date();
    var inRange = (now >= start && now <= end);
    var ocgName = "OCG_Normal";
    try {
      if(!inRange){
        app.alert(%s);
        return;  
      }
      if (typeof this.setOCGState === "function") {
        this.setOCGState(ocgName, inRange);
      } else if (typeof this.getOCGs === "function") {
        var ocgs = this.getOCGs();
        for (var i = 0; i < ocgs.length; i++) {
          var o = ocgs[i];
          if (o && o.name === ocgName) {
            o.state = inRange;
          }
        }
      }
    } catch (e) { app.alert(%s + ' ' + e); }
    } catch (e) { }
})();`, start.Format(time.RFC3339), end.Format(time.RFC3339), qExpired, qUnsupported)

	iref, err := ctx.IndRefForNewObject(types.Dict{
		"S":  types.Name("JavaScript"),
		"JS": types.StringLiteral(js),
	})
	if err != nil {
		fmt.Printf("injectOpenActionJS: %v\n", err)
	}

	ctx.RootDict["Names"] = types.Dict{
		"JavaScript": types.Dict{
			"Names": types.Array{
				types.StringLiteral("OpenActionJS"),
				*iref,
			},
		},
	}

	// Set OpenAction to the JavaScript action indirect reference.
	ctx.RootDict["OpenAction"] = *iref
}

func injectTimeJS(ctx *model.Context, start, end time.Time) {
	js := fmt.Sprintf(`(function () {
  try {
  app.alert("JS activated");
    var start = new Date("%s");
    var end   = new Date("%s");
    var now   = new Date();

    // Toggle global OCG visibility by flipping the state property on the named OCG.
    try {
      var ocgName = "OCG_Normal";
      if (typeof this.getOCGs === "function") {
        var ocgs = this.getOCGs();
        for (var i = 0; i < ocgs.length; i++) {
          var o = ocgs[i];
          if (o && o.name === ocgName) {
		  app.alert("Found OCG: " + ocgName);
            if (now >= start && now <= end) {
			app.alert("In range: enabling OCG");
              o.state = true;
            } else {
			 app.alert("Out of range: disabling OCG");
              o.state = true;
            }
          }
        }
      }
    } catch (e) { }

  } catch (e) { }
})();`, start.Format(time.RFC3339), end.Format(time.RFC3339))

	iref, err := ctx.IndRefForNewObject(types.Dict{
		"S":  types.Name("JavaScript"),
		"JS": types.StringLiteral(js),
	})
	if err != nil {
		fmt.Printf("injectTimeJS: %v\n", err)
	}
	ctx.RootDict["Names"] = types.Dict{
		"JavaScript": types.Dict{
			"Names": types.Array{
				types.StringLiteral("docjs"),
				*iref,
			},
		},
	}
}
