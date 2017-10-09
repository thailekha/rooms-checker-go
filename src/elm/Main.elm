module Main exposing (..)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing ( onClick, onInput )
import RemoteData exposing (WebData)
import Http
import Json.Encode as Encode
import Json.Decode as Decode
import Json.Decode.Pipeline as JDecode

-- component import example
import Components.Hello exposing ( hello )


-- APP
main : Program Never Model Msg
main =
  Html.program { init = init, view = view, update = update, subscriptions = (always Sub.none) }

type alias FreeRooms =
  {
  display : String
  }

-- MODEL
type alias Model = 
  { weekday : String
  , startTime : String
  , endTime : String
  , result : WebData (FreeRooms)
  }

init : ( Model, Cmd Msg )
init = 
  ({ weekday = "monday"
  , startTime = "9:15"
  , endTime = "9:15"
  , result = RemoteData.NotAsked
  }, Cmd.none)


-- UPDATE
type Msg 
  = NoOp 
  | SelectWeekday String
  | SelectStartTime String
  | SelectEndTime String
  | Submit
  | OnResponse (WebData (FreeRooms))

update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
  case msg of
    NoOp -> 
      ( model, Cmd.none  )
    SelectWeekday w ->
      ( {model | weekday = w}, Cmd.none )
    SelectStartTime s ->
      ( {model | startTime = s}, Cmd.none )
    SelectEndTime e ->
      ( {model | endTime = e}, Cmd.none )
    Submit ->
      ({ model | result = RemoteData.Loading }, send model )
    OnResponse response ->
      ({ model | result = response }, Cmd.none )


-- VIEW
-- Html is defined as: elem [ attribs ][ children ]
-- CSS can be applied via class names or inline style attrib
view : Model -> Html Msg
view model =
  div [ class "container", style [("margin-top", "30px"), ( "text-align", "center" )] ][    -- inline CSS (literal)
    label [] [text "Weekday"]
    , select [ onInput SelectWeekday ] [
      option [value "monday"] [ text "monday" ]
      , option [value "tuesday"] [ text "tuesday" ]
      , option [value "wednesday"] [ text "wednesday" ]
      , option [value "thursday"] [ text "thursday" ]
      , option [value "friday"] [ text "friday" ]
    ]
    , label [] [text "Start time"]
    , select [ onInput SelectStartTime ] [
      option [value "9:15"] [text "9:15"]
      , option [value "10:15"] [text "10:15"]
      , option [value "11:15"] [text "11:15"]
      , option [value "12:15"] [text "12:15"]
      , option [value "13:15"] [text "13:15"]
      , option [value "14:15"] [text "14:15"]
      , option [value "15:15"] [text "15:15"]
      , option [value "16:15"] [text "16:15"]
    ]
    , label [] [text "End time"]
    , select [ onInput SelectEndTime ] [
      option [value "9:15"] [text "9:15"]
      , option [value "10:15"] [text "10:15"]
      , option [value "11:15"] [text "11:15"]
      , option [value "12:15"] [text "12:15"]
      , option [value "13:15"] [text "13:15"]
      , option [value "14:15"] [text "14:15"]
      , option [value "15:15"] [text "15:15"]
      , option [value "16:15"] [text "16:15"]
    ]
    , br [] []
    , button [ onClick Submit ] [ text "Find" ]
    , br [] []
    , p [] [ maybeResult model.result ]
  ]

maybeResult : (WebData FreeRooms) -> Html Msg
maybeResult response =
    case response of
        RemoteData.NotAsked ->
            text "Look up something ..."
        RemoteData.Loading ->
            text "Loading..."
        RemoteData.Success freeRooms ->
            text freeRooms.display
        RemoteData.Failure error ->
            text (toString error)


send : Model -> Cmd Msg
send model = 
    Http.post ("http://localhost:3000/api/freetimes") (Http.jsonBody (encodeModel model)) freeRoomsDecoder
        |> RemoteData.sendRequest
        |> Cmd.map OnResponse


freeRoomsDecoder : Decode.Decoder FreeRooms
freeRoomsDecoder =
    JDecode.decode FreeRooms
        |> JDecode.required "FreeTimes" Decode.string


encodeModel : Model -> Encode.Value
encodeModel model =
    Encode.object
        [ ( "weekday", Encode.string model.weekday )
        , ( "startTime", Encode.string model.startTime )
        , ( "endTime", Encode.string model.endTime )
        ]


-- CSS STYLES
styles : { img : List ( String, String ) }
styles =
  {
    img =
      [ ( "width", "33%" )
      , ( "border", "4px solid #337AB7")
      ]
  }
