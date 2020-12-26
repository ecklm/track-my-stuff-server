function initMap() {
    const api_prefix = "/api/v1"
    const map = new google.maps.Map(document.getElementById("map"), {
        zoom: 14,
        center: { lat: 47.4725424, lng: 19.0437153 },
        mapTypeId: "terrain",
    });
    $.get(api_prefix + "/entity", function(entities, status){
      for (i in entities) {
          const entity = entities[i]
          $.get(api_prefix + "/record/" + entity.Name, function(data, status){
             flightPlanCoordinates = []
             for (i in data) {
                 flightPlanCoordinates.push({
                     lat: data[i].Position.Latitude,
                     lng: data[i].Position.Longitude,
                 });
             }
             const flightPath = new google.maps.Polyline({
                 path: flightPlanCoordinates,
                 geodesic: true,
                 strokeColor: "#FF0000",
                 strokeOpacity: 1.0,
                 strokeWeight: 2,
             });
             flightPath.setMap(map);
          })

          $.get(api_prefix + "/position/" + entity.Name, function(data, status){
             const newMarker = new google.maps.Marker({
                 position: {
                     lat: data.Position.Latitude,
                     lng: data.Position.Longitude,
                 },
                 map,
                 icon: entity.Icon,
                 title: entity.Name,
             });
          })
      }
      
    })
}
