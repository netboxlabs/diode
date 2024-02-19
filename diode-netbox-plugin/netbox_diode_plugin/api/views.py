from django.contrib.contenttypes.models import ContentType
from django.db.models import QuerySet
from extras.models import CachedValue
from netbox.search import LookupTypes
from netbox.search.backends import search_backend
from rest_framework import generics

from netbox_diode_plugin.api.serializers import CachedValueSerializer


class ObjectStateList(generics.ListAPIView):
    serializer_class = CachedValueSerializer
    authentication_classes = []  # disables authentication
    permission_classes = []
    queryset = CachedValue.objects.all()

    def get_queryset(self):

        # Restrict results by object type
        # object_types = []
        # for obj_type in self.request.query_params.getlist('obj_types', []):
        #     app_label, model_name = obj_type.split('.')
        #     object_types.append(ContentType.objects.get_by_natural_key(app_label, model_name))

        # Restrict results by object type
        # obj_type.split('.')[0] = app_label // obj_type.split('.')[1] = model_name
        object_types = [ContentType.objects.get_by_natural_key(obj_type.split('.')[0], obj_type.split('.')[1]) for
                        obj_type in self.request.query_params.getlist('obj_types', [])]

        print(object_types)

        lookup = self.request.query_params.get('lookup') or LookupTypes.PARTIAL
        results = search_backend.search(
            self.request.query_params.get('q'),
            # user=self.request.user,
            object_types=object_types,
            lookup=lookup
        )
        print(len(results))
        print(results)

        assert self.queryset is not None, (
                "'%s' should either include a `queryset` attribute, "
                "or override the `get_queryset()` method."
                % self.__class__.__name__
        )

        queryset = results

        if isinstance(queryset, QuerySet):
            # Ensure queryset is re-evaluated on each request.
            queryset = queryset.all()
        return queryset
