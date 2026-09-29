package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func classifyRoot() map[string]any {
	return map[string]any{
		"inputs": map[string]any{
			"classify": map[string]any{
				"details": []any{
					map[string]any{"classified": "pedestrian"},
					map[string]any{"classified": "car"},
				},
			},
		},
	}
}

// A condition read back from MongoDB decodes its `any` operand as primitive.A,
// not []any; the `in` operator must still match it.
func TestEvaluateConditionInOperandDecodedFromBSON(t *testing.T) {
	stored := StageCondition{
		Path:  "inputs.classify.details.*.classified",
		Op:    ConditionOpIn,
		Value: []any{"lorry", "car"},
	}
	raw, err := bson.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	var decoded StageCondition
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded.Value.(primitive.A); !ok {
		t.Fatalf("expected primitive.A operand after BSON decode, got %T", decoded.Value)
	}

	if !EvaluateCondition(&decoded, classifyRoot()) {
		t.Fatal("in condition with a BSON-decoded operand should match a detected car")
	}

	decoded.Value = primitive.A{"lorry", "cyclist"}
	if EvaluateCondition(&decoded, classifyRoot()) {
		t.Fatal("in condition should not match when no listed class was detected")
	}
}

func TestEvaluateConditionAcceptsTypedCollections(t *testing.T) {
	in := &StageCondition{Path: "inputs.classify.details.*.classified", Op: ConditionOpIn, Value: []string{"car"}}
	if !EvaluateCondition(in, classifyRoot()) {
		t.Fatal("in should accept a []string operand")
	}

	bsonRoot := map[string]any{
		"inputs": primitive.M{
			"classify": primitive.M{
				"details": primitive.A{primitive.M{"classified": "car"}},
				"labels":  []string{"car", "lorry"},
			},
		},
	}
	if !EvaluateCondition(&StageCondition{Path: "inputs.classify.details.*.classified", Op: ConditionOpEq, Value: "car"}, bsonRoot) {
		t.Fatal("path resolution should traverse primitive.M and primitive.A")
	}
	if !EvaluateCondition(&StageCondition{Path: "inputs.classify.labels", Op: ConditionOpContains, Value: "lorry"}, bsonRoot) {
		t.Fatal("contains should accept a typed slice candidate")
	}
	if EvaluateCondition(&StageCondition{Path: "inputs.classify.details.*.classified", Op: ConditionOpIn, Value: []byte("car")}, bsonRoot) {
		t.Fatal("a []byte operand is not a list")
	}
}
