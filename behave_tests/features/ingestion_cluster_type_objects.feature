Feature: Tests for ingestion of cluster type
    Validate the behavior of the ingestion of cluster type

@smoke
@ingestion.cluster_type
Scenario: Ingestion of new cluster type
    Given a new cluster type "VMWare"
    When the cluster type is ingested
    Then the cluster type is created in the database

@smoke
@ingestion.cluster_type
Scenario: Ingestion of existing cluster type
    Given cluster type "VMWare" already exists in the database
    When the cluster type is ingested
    Then the cluster type remains the same


@smoke
@ingestion.cluster_type
Scenario: Ingestion of cluster type to update its description
    Given cluster type "VMWare" with description "some string"
    When the cluster type is ingested with the updates
    Then the cluster type is updated in the database
