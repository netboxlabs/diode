Feature: Tests for ingestion of cluster group
    Validate the behavior of the ingestion of cluster group

@smoke
@ingestion.cluster_group
Scenario: Ingestion of new cluster group
    Given a new cluster group "North America"
    When the cluster group is ingested
    Then the cluster group is created in the database

@smoke
@ingestion.cluster_group
Scenario: Ingestion of existing cluster group
    Given cluster group "North America" already exists in the database
    When the cluster group is ingested
    Then the cluster group remains the same


@smoke
@ingestion.cluster_group
Scenario: Ingestion of cluster group to update its description
    Given cluster group "North America" with description "some string"
    When the cluster group is ingested with the updates
    Then the cluster group is updated in the database
