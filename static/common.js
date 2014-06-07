var loltools = (function() {
  var module = {}
  
  module.registerForm = function(formId, submitUrl) {
    var req;
    $("form#"+formId).submit(function(event){
      event.preventDefault();
      if (req) {
        req.abort();
      }
      var $errors = $("#errors")
      var $form = $(this);
      var $inputs = $.merge(
        $form.find("input, select, button, textarea"),
        $("input[form='"+formId+"'], select[form='"+formId+"'],"+
          "button[form='"+formId+"'], textarea[form='"+formId+"']"));
      var $submit = $.merge(
        $form.find("#submit"),
        $("input[type='submit'][form='"+formId+"']"));

      var data = $form.serialize()
      
      $errors.empty();
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
      });
    });
  }
  
  return module;
})();