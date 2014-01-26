var loltools = (function() {
  var module = {}
  
  module.registerForm = function(formId, submitUrl) {
    var req;
    $("#"+formId).submit(function(event){
      if (req) {
        req.abort();
      }
      var $errors = $("#errors")
      var $form = $(this);
      var $inputs = $form.find("input, select, button, textarea");
      var $result = $form.find("#result");
      var $submit = $form.find("#submit");

      var data = $form.serialize()

      var originalSubmitVal = $submit.val()
      
      $errors.empty();
      $submit.val("sending...");
      $inputs.prop("disabled", true);
      
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
        } else if (jqXHR.status == 204) {
          window.location.reload()
        }
      });
      req.fail(function (jqXHR, textStatus, errorThrown){
        $errors.append("<div>error: " + errorThrown + ": " + jqXHR.responseText + "</div>");
      });
      req.always(function () {
        $inputs.prop("disabled", false);
        $submit.val(originalSubmitVal);
      });
      event.preventDefault();
    });
  }
  
  return module;
})();