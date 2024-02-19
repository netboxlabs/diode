from django.urls import path, include

from netbox.api.routers import NetBoxRouter

from .views import ObjectStateList

# app = "netbox_diode_plugin"

router = NetBoxRouter()

urlpatterns = [
    path('object-state/', ObjectStateList.as_view()),
    path('', include(router.urls))
]
