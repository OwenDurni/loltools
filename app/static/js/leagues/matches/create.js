// Extend loltools module with functions specific to /leagues/matches/create.html
var loltools = (function(loltools) {
  loltools.pairingsAddRow = function(buttonSel) {
    var $button = $(buttonSel);
    $button.click(function (event) {
      event.preventDefault();
      var $teamsDropdownTemplate = $("#template-team-list");
      var $pairingsTable = $("#pairings");
      var $pairingsLastRow = $pairingsTable.find("tr").last();
    
      var $home =
        $teamsDropdownTemplate.clone()
          .removeAttr("id").removeAttr("style")
          .attr("name", "home-team");
      var $away =
        $teamsDropdownTemplate.clone()
          .removeAttr("id").removeAttr("style")
          .attr("name", "away-team");
          
      $pairingsLastRow.before(
        $(document.createElement("tr")).append(
          $(document.createElement("td")).append($home),
          $(document.createElement("td")).append($away)));
    });
  }
  
  loltools.pairingsDelRow = function(buttonSel) {
    var $button = $(buttonSel);
    $button.click(function (event) {
      event.preventDefault();
      var $teamsDropdownTemplate = $("#template-team-list");
      var $pairingsTable = $("#pairings");
      var $rows = $pairingsTable.find("tr");
      
      if ($rows.size() <= 2) {
        // Do not remove if only rows are the header and the row with buttons.
        return;
      }
      $rows.get($rows.size()-2).remove();
    });
  }
  
  return loltools;
}(loltools));