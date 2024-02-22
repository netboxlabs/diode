from extras.models import CachedValue, ObjectChange
from rest_framework import serializers
from utilities.api import get_serializer_for_model


class ObjectStateSerializer(serializers.Serializer):
    object_type = serializers.SerializerMethodField(read_only=True)
    object_change_id = serializers.SerializerMethodField(read_only=True)
    object = serializers.SerializerMethodField(read_only=True)

    # class Meta:
    #     model = CachedValue
    #     fields = ['object_type', 'object_change', 'object']

    def get_object_type(self, instance):
        return self.context.get('object_type')

    def get_object_change_id(self, instance):
        object_changed = ObjectChange.objects.filter(changed_object_id=instance.id).values_list('id', flat=True).latest('id')
        return object_changed

    def get_object(self, instance):
        serializer = get_serializer_for_model(instance)

        object_data = instance.__class__.objects.filter(id=instance.id)
        print(object_data)

        context = {'request': self.context.get('request')}
        return serializer(object_data, context=context, many=True).data[0]