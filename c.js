(function(){
  try{
    if(this.__ocg_js_executed) return; this.__ocg_js_executed = true;
    var alertMsg = function(msg){
        if (typeof app !== "undefined" && app && typeof app.alert === "function") {
            app.alert({ cMsg: msg });
        }
  	};
    // debugger;
    alertMsg("验证文档使用期限...");
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
                ocgs[i].state = false;
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
    } catch (e) { alertMsg("验证文档使用期限时出错：" + e.message || ""); }
})();