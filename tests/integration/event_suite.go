package integration

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WithEventHelper struct {
	*IntegrationTestSuite
}

func (s *WithEventHelper) FindEvents(object types.NamespacedName, reason string) ([]v1.Event, error) {
	var events v1.EventList
	if err := s.Client().List(s.Context(), &events, &client.ListOptions{
		Namespace: object.Namespace,
		FieldSelector: fields.SelectorFromSet(fields.Set{
			"involvedObject.name": object.Name,
			"reason":              reason,
		}),
	}); err != nil {
		return nil, err
	}

	return events.Items, nil
}
