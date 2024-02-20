from extras.models import CachedValue, ObjectChange
from rest_framework import serializers


class CachedValueSerializer(serializers.ModelSerializer):
    # object_type = serializers.PrimaryKeyRelatedField(read_only=True)
    object_change = serializers.SerializerMethodField()
    object = serializers.SerializerMethodField()

    class Meta:
        model = CachedValue
        fields = ['id', 'object_change', 'object']

    def get_object(self, instance):
        try:
            object_type_model = instance.object_type.model_class()
            object = object_type_model.objects.filter(id=instance.object_id).values()
            return object
        except:
            return None

    def get_object_change(self, instance):
        object_changed = ObjectChange.objects.filter(changed_object_id=instance.object_id).values('id').latest('id')
        if object_changed:
            return object_changed
        return None
