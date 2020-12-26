package main

import (
    "os"
    "log"
    "fmt"
    "time"
    "context"
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "cloud.google.com/go/firestore"
    "google.golang.org/api/iterator"
)

func createClient(firestore_context context.Context) *firestore.Client {
        // Sets your Google Cloud Platform project ID.
        projectID := os.Getenv("GCP_PROJECT_ID")

        firestore_client, err := firestore.NewClient(firestore_context, projectID)
        if err != nil {
                log.Fatalf("Failed to create firestore_client: %v", err)
        }
        // Close firestore_client when done with
        // defer firestore_client.Close()
        return firestore_client
}

var (
    firestore_context context.Context
    firestore_client *firestore.Client
    firestore_collection = map[string]string{
        "records": "track-records",
        "positions": "track-positions",
        "entities": "track-entities",
    }
)

func init() {
    firestore_context = context.Background()
    firestore_client = createClient(firestore_context)
}

func main() {
    defer firestore_client.Close()

    e := echo.New()

    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    e.Static("/map", "map")

    e.GET("/records/:entity", getRecords)
    e.POST("/records/:entity", addRecord)

    e.GET("/position/:entity", getPosition)

    e.GET("/entity", listEntities)

    e.GET("/api/proxy/maps.api.js", func(c echo.Context)error{
        url := fmt.Sprintf(
            "https://maps.googleapis.com/maps/api/js?key=%s&callback=initMap&libraries=&v=weekly",
            os.Getenv("GOOGLE_MAPS_API_KEY"))
        return c.Redirect(http.StatusMovedPermanently, url)
    })

    e.Logger.Fatal(e.Start(":8080"))
}

type Position struct {
    Longitude float32 `json:"longitude" form:"longitude" query:"longitude"`
    Latitude float32 `json:"latitude" form:"latitude" query:"latitude"`
}

type Record struct {
    Entity string
    Position Position
    Time time.Time
}

func addRecord(c echo.Context) error {
    position := new(Position)
    if err := c.Bind(position); err != nil {
        return err
    }
    // TODO: Check if both are provided

    new_record := Record{c.Param("entity"), *position, time.Now()}

    _, _, err := firestore_client.Collection(firestore_collection["records"]).
        Add(firestore_context, new_record)
    if err != nil {
        return c.JSON(http.StatusInternalServerError,
                      map[string]interface{}{
                         "Reason": fmt.Sprintf("Failed adding record: %v", err),
                      })
    }
    if err := setPosition(c, new_record); err != nil {
        return err
    }
    return c.JSON(http.StatusOK, new_record.Position)
}

func getRecords(c echo.Context) error {
    entity := c.Param("entity")

    var results []map[string]interface{}

    iter := firestore_client.Collection(firestore_collection["records"]).
        // Where and OrderBy doesn't work combined here for some reason...
        // Where("Entity", "==", entity).
        OrderBy("Time", firestore.Desc).
        Documents(firestore_context)
    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return c.JSON(http.StatusInternalServerError,
                          map[string]interface{}{"Reason": err},
            )
        }
        data := doc.Data()
        if (data["Entity"] == entity) {
            results = append(results, data)
        }
    }
    return c.JSON(http.StatusOK, results)
}

func setPosition(c echo.Context, new_record Record) error {
    _, err := firestore_client.Collection(firestore_collection["positions"]).
        Doc(new_record.Entity).Set(firestore_context, new_record)
    if err != nil {
        return c.JSON(http.StatusInternalServerError,
                      map[string]interface{}{"Reason": err})
    }
    return nil
}

func getPosition(c echo.Context) error {
    entity := c.Param("entity")

    doc, err := firestore_client.Collection(firestore_collection["positions"]).
        Doc(entity).Get(firestore_context)
    if err != nil {
        return c.JSON(http.StatusInternalServerError,
                      map[string]interface{}{"Reason": err})
    }
    return c.JSON(http.StatusOK, doc.Data())
}

func listEntities(c echo.Context) error {
    entities := []map[string]interface{}{}

    iter := firestore_client.Collection(firestore_collection["entities"]).
        Documents(firestore_context)
    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return c.JSON(http.StatusInternalServerError,
                          map[string]interface{}{"Reason": err},
            )
        }
        entity := doc.Data()
        entities = append(entities, entity)
    }
    return c.JSON(http.StatusOK, entities)
}
