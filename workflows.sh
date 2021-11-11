GRPC_WORKFLOW_TEMPLATE=$(cat .github/grpc-template.yaml)
GIN_WORKFLOW_TEMPLATE=$(cat .github/gin-template.yaml)
RUST_WORKFLOW_TEMPLATE=$(cat .github/rust-template.yaml)

for ROUTE in $(cat .github/grpc-ci-paths); do
    echo "generating workflow for ${ROUTE}"

    WORKFLOW=$(echo "${GRPC_WORKFLOW_TEMPLATE}" | sed "s/{{ROUTE}}/${ROUTE}/g")

    echo "${WORKFLOW}" > .github/workflows/${ROUTE}.yaml
done

for ROUTE in $(cat .github/gin-ci-paths); do
    echo "generating workflow for ${ROUTE}"

    WORKFLOW=$(echo "${GIN_WORKFLOW_TEMPLATE}" | sed "s/{{ROUTE}}/${ROUTE}/g")

    echo "${WORKFLOW}" > .github/workflows/${ROUTE}.yaml
done

for ROUTE in $(cat .github/rust-ci-paths); do
    echo "generating workflow for ${ROUTE}"

    WORKFLOW=$(echo "${RUST_WORKFLOW_TEMPLATE}" | sed "s/{{ROUTE}}/${ROUTE}/g")

    echo "${WORKFLOW}" > .github/workflows/${ROUTE}.yaml
done