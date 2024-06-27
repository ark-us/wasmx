package types

var (
	// nonconsensusless
	HOOK_START_NODE = "StartNode"
	HOOK_SETUP_NODE = "SetupNode"

	// consenssus
	HOOK_BEGIN_BLOCK      = "BeginBlock"
	HOOK_END_BLOCK        = "EndBlock"
	HOOK_CREATE_VALIDATOR = "CreatedValidator"

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
	SourceModule  string   `json:"sourceModule"`
	TargetModules []string `json:"targetModules"`
}

var DEFAULT_HOOKS = []Hook{
	{
		Name:          HOOK_BEGIN_BLOCK,
		SourceModule:  ROLE_CONSENSUS,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          HOOK_END_BLOCK,
		SourceModule:  ROLE_CONSENSUS,
		TargetModules: []string{ROLE_GOVERNANCE, ROLE_DISTRIBUTION},
	},
	{
		Name:          HOOK_CREATE_VALIDATOR,
		SourceModule:  ROLE_CONSENSUS,
		TargetModules: []string{},
	},
	{
		Name:          AfterValidatorCreated,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterValidatorBonded,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterValidatorRemoved,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterValidatorBeginUnbonding,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterDelegationModified,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          AfterUnbondingInitiated,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeValidatorModified,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeDelegationCreated,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeDelegationSharesModified,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeDelegationRemoved,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
	{
		Name:          BeforeValidatorSlashed,
		SourceModule:  ROLE_STAKING,
		TargetModules: []string{ROLE_SLASHING},
	},
}

var DEFAULT_HOOKS_NONC = []Hook{
	{
		Name:          HOOK_START_NODE,
		SourceModule:  ROLE_HOOKS_NONC,
		TargetModules: []string{ROLE_CONSENSUS, ROLE_CHAT},
	},
	{
		Name:          HOOK_SETUP_NODE,
		SourceModule:  ROLE_HOOKS_NONC,
		TargetModules: []string{ROLE_CONSENSUS},
	},
}
