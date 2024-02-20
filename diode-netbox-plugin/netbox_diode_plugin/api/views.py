from django.contrib.contenttypes.models import ContentType
from extras.models import CachedValue
from netbox.search import LookupTypes
from netbox.search.backends import search_backend
from rest_framework import generics
from rest_framework.exceptions import ValidationError

from netbox_diode_plugin.api.serializers import CachedValueSerializer


class ObjectStateList(generics.ListAPIView):
    serializer_class = CachedValueSerializer
    authentication_classes = []  # disables authentication
    permission_classes = []
    queryset = CachedValue.objects.all()

    def get_queryset(self):

        '''
        Search for objects in the cache.
        If obj_type parameter is not in the parameters, raise an ValidationError.
        When object ID is provided in the request, make a search using it and object_type to retrieve only one object.
        If ID is not provided, uses q parameter to find one or more objects.
        The lookup is locked to use "iexact".
        '''

        obj_type = self.request.query_params.get('obj_type', None)

        if not obj_type:
            raise ValidationError('obj_type parameter is required')

        app_label, model_name = obj_type.split('.')
        object_types = [ContentType.objects.get_by_natural_key(app_label, model_name)]
        object_id = self.request.query_params.get('id', None)

        if object_id:
            # field is to avoid duplicated results.
            queryset = CachedValue.objects.filter(object_id=object_id).filter(object_type__in=object_types).filter(
                field="name")
        else:
            search_value = self.request.query_params.get('q', None)
            if not search_value:
                raise ValidationError('id or q parameter is required')
            lookup = LookupTypes.EXACT
            queryset = search_backend.search(
                search_value,
                # user=self.request.user,
                object_types=object_types,
                lookup=lookup
            )

        return queryset
