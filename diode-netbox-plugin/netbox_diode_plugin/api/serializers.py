from extras.models import CachedValue
from rest_framework import serializers


class CachedValueSerializer(serializers.ModelSerializer):
    class Meta:
        model = CachedValue
        fields = '__all__'
