package types

var (
	// nonconsensusless
	HOOK_START_NODE   = "StartNode"
	HOOK_SETUP_NODE   = "SetupNode"
	HOOK_NEW_SUBCHAIN = "NewSubChain"

	// consenssus
	HOOK_BEGIN_BLOCK      = "BeginBlock"
	HOOK_END_BLOCK        = "EndBlock"
	HOOK_CREATE_VALIDATOR = "CreatedValidator"
	HOOK_ROLE_CHANGED     = "RoleChanged"

	// staking
	AfterValidatorCreated          = "AfterValidatorCreated"
	AfterValidatorBonded           = "AfterValidatorBonded"
	AfterValidatorRemoved          = "AfterValidatorRemoved"
	AfterValidatorBeginUnbonding   = "AfterValidatorBeginUnbonding"
	AfterDelegationModified        = "AfterDelegationModified"
	AfterUnbondingInitiated        = "AfterUnbondingInitiated"
	BeforeValidatorModified        = "BeforeValidatorModified"
	BeforeDelegationCreated        = "BeforeDelegationCreated"
	BeforeDelegationSharesModified = "BeforeDelegationSharesModified"
	BeforeDelegationRemoved        = "BeforeDelegationRemoved"
	BeforeValidatorSlashed         = "BeforeValidatorSlashed"
)

type Hook struct {
	Name          string   `json:"name"`
	SourceModules []string `json:"sourceModules"`
	TargetModules []string `json:"targetModules"`
}

var DEFAULT_HOOKS = []Hook{
	{
		Name:          HOOK_BEGIN_BLOCK,
		SourceModules: []string{ROLE_CONSENSUS},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          HOOK_END_BLOCK,
		SourceModules: []string{ROLE_CONSENSUS},
		TargetModules: []string{ROLE_GOVERNANCE, ROLE_DISTRIBUTION},
	},
	{
		Name:          HOOK_CREATE_VALIDATOR,
		SourceModules: []string{ROLE_CONSENSUS},
		TargetModules: []string{},
	},
	{
		Name:          AfterValidatorCreated,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterValidatorBonded,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterValidatorRemoved,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterValidatorBeginUnbonding,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterDelegationModified,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterUnbondingInitiated,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeValidatorModified,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeDelegationCreated,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeDelegationSharesModified,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeDelegationRemoved,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeValidatorSlashed,
		SourceModules: []string{ROLE_STAKING},
		TargetModules: []string{ROLE_SLASHING},
	},
}

var DEFAULT_HOOKS_NONC = []Hook{
	{
		Name:          HOOK_START_NODE,
		SourceModules: []string{ROLE_HOOKS_NONC},
		TargetModules: []string{ROLE_CONSENSUS, ROLE_MULTICHAIN_REGISTRY_LOCAL, ROLE_CHAT},
	},
	{
		Name:          HOOK_SETUP_NODE,
		SourceModules: []string{ROLE_HOOKS_NONC},
		TargetModules: []string{ROLE_CONSENSUS, ROLE_LOBBY},
	},
	{
		Name:          HOOK_NEW_SUBCHAIN,
		SourceModules: []string{ROLE_HOOKS_NONC},
		TargetModules: []string{ROLE_METAREGISTRY, ROLE_MULTICHAIN_REGISTRY_LOCAL},
	},
}
