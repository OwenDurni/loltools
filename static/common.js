var loltools = (function() {
  var module = {}
  
  module.registerForm = function(formId, submitUrl) {
    var req;
    $("#"+formId).submit(function(event){
      if (req) {
        req.abort();
      }
      var $form = $(this);
      var $inputs = $form.find("input, select, button, textarea");
      var $result = $form.find("#result");

      var data = $form.serialize()

      $inputs.prop("disabled", true);
      $result.text("sending...");
      
      req = $.ajax({
        url: submitUrl,
        type: "post",
        data: data
      });

      req.done(function (response, textStatus, jqXHR){
        if (jqXHR.status == 201) {
          var loc = jqXHR.getResponseHeader("Location");
          if (loc) {
            // Redirect to created resource.
            window.location.href = loc;
          }
        }
        $result.text("sent: " + jqXHR.status);
      });
      req.fail(function (jqXHR, textStatus, errorThrown){
        $result.text("error: " + errorThrown + ": " + jqXHR.responseText);
      });
      req.always(function () {
        $inputs.prop("disabled", false);
      });
      event.preventDefault();
    });
  }
  
  return module;
})();