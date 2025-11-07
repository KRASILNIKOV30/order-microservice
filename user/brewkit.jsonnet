local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'user',
];

local proto = [
    'api/client/testinternal/testinternal.proto',
    'api/server/userinternal/userinternal.proto',
];

project.project(appIDs, proto)