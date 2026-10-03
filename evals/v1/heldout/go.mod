// These fixtures belong to separate pinned upstream modules. The evaluator
// copies one test into its matching candidate; root Go tests must not try to
// compile this mixed-package fixture directory as a Fabric package.
module fabric.local/heldout-fixtures

go 1.27.1
