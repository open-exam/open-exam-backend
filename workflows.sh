WORKFLOW_TEMPLATE=$(cat .github/grpc-template.yaml)

# iterate each route in routes directory
for ROUTE in $(cat .github/grpc-ci-paths); do
    echo "generating workflow for routes/${ROUTE}"

    # replace template route placeholder with route name
    WORKFLOW=$(echo "${WORKFLOW_TEMPLATE}" | sed "s/{{ROUTE}}/${ROUTE}/g")

    # save workflow to .github/workflows/{ROUTE}
    echo "${WORKFLOW}" > .github/workflows/${ROUTE}.yaml
done