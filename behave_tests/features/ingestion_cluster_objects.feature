Feature: Tests for ingestion of cluster
    Validate the behavior of the ingestion of cluster

@smoke
@ingestion.cluster
Scenario: Ingestion of new cluster
    Given a new cluster "aws-us-east-1"
    When the cluster is ingested
    Then the cluster, "undefined" group, "undefined" type and "undefined" site are created in the database

@smoke
@ingestion.cluster
Scenario: Ingestion of existing cluster
    Given cluster "aws-us-east-1" already exists in the database
    When the cluster is ingested
    Then the cluster remains the same

@smoke
@ingestion.cluster
Scenario: Ingestion of cluster object to update the group, type site and description
    Given cluster "aws-us-east-1" with group "NA", type "AWS", site "NA-NY" and description "some string"
    Then check if the group "NA", type "AWS" and site "NA-NY" exist in the database and remove them
    When the cluster object is ingested with the updates
    Then the group "NA", type "AWS" and site "NA-NY" are created and the cluster updated
