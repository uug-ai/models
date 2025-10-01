package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type AnalysisWrapper struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Key                string             `json:"key,omitempty"`
	Provider           string             `json:"provider" bson:"provider"`
	Source             string             `json:"source" bson:"source"`
	InProcess          int64              `json:"in_process,omitempty"`
	Timestamp          int64              `json:"timestamp,omitempty"`
	Start              int64              `json:"start,omitempty"`
	End                int64              `json:"end,omitempty"`
	UserId             string             `json:"user_id,omitempty"`
	DeviceId           string             `json:"device_id,omitempty"`
	AsyncOperations    []string           `json:"asyncOperations,omitempty"`
	RequiredOperations []string           `json:"requiredOperations,omitempty"`
	ResolvedOperations []string           `json:"resolvedOperations,omitempty"`
	Favourite          bool               `json:"favourite,omitempty"`
	Data               Analysis           `json:"data,omitempty"`
}

type AnalysisShort struct {
	Id                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Key               string             `json:"key,omitempty"`
	Provider          string             `json:"provider" bson:"provider"`
	Source            string             `json:"source" bson:"source"`
	FileUrl           string             `json:"src" bson:"src"`
	Timestamp         int64              `json:"timestamp,omitempty"`
	Start             int64              `json:"start,omitempty"`
	UserId            string             `json:"user_id,omitempty"`
	DeviceId          string             `json:"device_id,omitempty"`
	ThumbnaiFile      string             `json:"thumbnail_file,omitempty"`
	ThumbnailProvider string             `json:"thumbnail_provider,omitempty"`
	ThumbnailUrl      string             `json:"thumbnail_src"`
	Favourite         bool               `json:"favourite,omitempty"`
	Data              *Analysis          `json:"data,omitempty"`
}

type Analysis struct {
	Classify      Classify        `json:"classify" bson:"classify"`
	Counting      Counting        `json:"counting" bson:"counting"`
	DominantColor DominantColor   `json:"dominantcolor" bson:"dominantcolor"`
	Thumby        Thumby          `json:"thumby" bson:"thumby"`
	Sprite        Sprite          `json:"sprite" bson:"sprite"`
	FaceRedaction []FaceRedaction `json:"faceRedaction" bson:"faceRedaction"`
}

type Classify struct {
	Properties []string          `json:"properties" bson:"properties"`
	Details    []ClassifyDetails `json:"details" bson:"details"`
}

type ClassifyDetails struct {
	Id               string      `json:"id" bson:"id"`
	X                float64     `json:"x" bson:"x"`
	Y                float64     `json:"y" bson:"y"`
	W                float64     `json:"w" bson:"w"`
	H                float64     `json:"h" bson:"h"`
	FrameWidth       float64     `json:"frameWidth" bson:"frameWidth"`
	FrameHeight      float64     `json:"frameHeight" bson:"frameHeight"`
	MeanColor        float64     `json:"meanColor" bson:"meanColor"`
	Frame            float64     `json:"frame" bson:"frame"`
	Occurence        float64     `json:"occurence" bson:"occurence"`
	Distance         float64     `json:"distance" bson:"distance"`
	Valid            bool        `json:"valid" bson:"valid"`
	Classified       string      `json:"classified" bson:"classified"`
	Traject          [][]float64 `json:"traject" bson:"traject"`
	ColorString      []string    `json:"colorStr" bson:"colorStr"`
	IsStatic         bool        `json:"isStatic" bson:"isStatic"`
	Frames           []int64     `json:"frames" bson:"frames"`
	StaticDistance   float64     `json:"staticDistance" bson:"staticDistance"`
	TrajectCentroids [][]float64 `json:"trajectCentroids" bson:"trajectCentroids"`
}

type Color struct {
	Name  string `json:"name" bson:"name"`
	Count int    `json:"count" bson:"count"`
}

type Counting struct {
	Details []CountingDetail `json:"detail,omitempty"`
	Regions []Region         `json:"regions,omitempty"`
	Records []CountingRecord `json:"records" bson:"records"`
}

type CountingDetail struct {
	DeviceID    string    `json:"deviceId" bson:"deviceId"`
	ObjectName  string    `json:"objectName" bson:"objectName"`
	Timestamp   string    `json:"timestamp" bson:"timestamp"`
	VideoWidth  float64   `json:"videoWidth" bson:"videoWidth"`
	VideoHeight float64   `json:"videoHeight" bson:"videoHeight"`
	Segment     string    `json:"segment" bson:"segment"`
	ObjectId    string    `json:"objectId" bson:"objectId"`
	Position    []float64 `json:"position" bson:"position"`
}

type CountingRecord struct {
	Type       string  `json:"type" bson:"type"`
	SegmentId  string  `json:"segmentId" bson:"segmentId"`
	Username   string  `json:"username" bson:"username"`
	AlertId    string  `json:"alert_id" bson:"alert_id"`
	AlertName  string  `json:"alert_name" bson:"alert_name"`
	Timestamp  string  `json:"timestamp" bson:"timestamp"`
	DeviceId   string  `json:"deviceId" bson:"deviceId"`
	ObjectName string  `json:"objectName" bson:"objectName"`
	Count      int     `json:"count" bson:"count"`
	Duration   float64 `json:"duration" bson:"duration"`
}

type DominantColor struct {
	Rgbs [][]float64 `json:"rgbs" bson:"rgbs"`
	Hexs []string    `json:"hexs" bson:"hexs"`
}

type Thumby struct {
	ThumbnailFile string `json:"filename" bson:"filename"`
	Provider      string `json:"provider" bson:"provider"`
	Base64        string `json:"base64" bson:"base64"`
	Interval      int    `json:"interval" bson:"interval"` // Only for sprites (Hacky)
	//Quality       string `json:"quality" bson:"quality"` -> Is getting type errors due to new mongodb driver.. need to fix.
	//Width  string `json:"width" bson:"width"` -> Is getting type errors due to new mongodb driver.. need to fix.
	//Height string `json:"height" bson:"height"` -> Is getting type errors due to new mongodb driver.. need to fix.
}

type Sprite struct {
	SpriteFile string `json:"filename" bson:"filename"`
	Provider   string `json:"provider" bson:"provider"`
	Interval   int    `json:"interval" bson:"interval"`
}

type AnalysisFilter struct {
	Start      int64    `json:"start" bson:"start"`
	Limit      int64    `json:"limit" bson:"limit"`
	Sort       string   `json:"sort" bson:"sort"`
	Operations []string `json:"operations" bson:"operations"`
}

type Thumbnail struct {
	Key           string `json:"key" bson:"key"`
	ThumbnailFile string `json:"thumbnailFile" bson:"thumbnailFile"`
}

type FaceRedactionTrack struct {
	Id               string             `json:"id" bson:"id"`
	Classified       string             `json:"classified" bson:"classified"`
	Frames           []int64            `json:"frames" bson:"frames"`
	Traject          [][]float64        `json:"traject" bson:"traject"` // [x1, y1, x2, y2, frame]
	ColorString      []string           `json:"colorStr,omitempty" bson:"colorStr,omitempty"`
	Selected         bool               `json:"selected" bson:"selected"`
	DeletedFrames    []int64            `json:"deletedFrames,omitempty" bson:"deletedFrames,omitempty"`
	FrameCoordinates map[int64]TrackBox `json:"frameCoordinates,omitempty" bson:"frameCoordinates,omitempty"` // frame -> [x1, y1, x2, y2]
}

type TrackBox struct {
	X1       float64 `json:"x1" bson:"x1"`
	Y1       float64 `json:"y1" bson:"y1"`
	X2       float64 `json:"x2" bson:"x2"`
	Y2       float64 `json:"y2" bson:"y2"`
	TrackId  string  `json:"trackId,omitempty" bson:"trackId,omitempty"`
	Smoothed bool    `json:"smoothed" bson:"smoothed"`
	Edited   bool    `json:"edited" bson:"edited"`
}

type FaceRedaction struct {
	Id     primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Tracks []FaceRedactionTrack `json:"tracks" bson:"tracks"`
}
