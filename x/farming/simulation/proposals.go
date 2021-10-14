package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/tendermint/farming/app/params"
	"github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"
)

// Simulation operation weights constants.
const (
	OpWeightSimulateAddPublicPlanProposal    = "op_weight_add_public_plan_proposal"
	OpWeightSimulateUpdatePublicPlanProposal = "op_weight_update_public_plan_proposal"
	OpWeightSimulateDeletePublicPlanProposal = "op_weight_delete_public_plan_proposal"
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSimulateAddPublicPlanProposal,
			params.DefaultWeightAddPublicPlanProposal,
			SimulateAddPublicPlanProposal(ak, bk, k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateUpdatePublicPlanProposal,
			params.DefaultWeightUpdatePublicPlanProposal,
			SimulateUpdatePublicPlanProposal(ak, bk, k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateDeletePublicPlanProposal,
			params.DefaultWeightDeletePublicPlanProposal,
			SimulateDeletePublicPlanProposal(ak, bk, k),
		),
	}
}

// SimulateAddPublicPlanProposal generates random public plan proposal content
func SimulateAddPublicPlanProposal(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		params := k.GetParams(ctx)
		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return nil
		}

		poolCoins, err := mintPoolCoins(ctx, r, bk, simAccount)
		if err != nil {
			return nil
		}

		addRequests := genAddRequestProposals(r, ctx, simAccount, poolCoins)

		return types.NewPublicPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			addRequests,
			[]*types.UpdateRequestProposal{},
			[]*types.DeleteRequestProposal{},
		)
	}
}

// SimulateUpdatePublicPlanProposal generates random public plan proposal content
func SimulateUpdatePublicPlanProposal(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		params := k.GetParams(ctx)
		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return nil
		}

		poolCoins, err := mintPoolCoins(ctx, r, bk, simAccount)
		if err != nil {
			return nil
		}

		req := &types.UpdateRequestProposal{}

		plans := k.GetPlans(ctx)
		for _, p := range plans {
			if p.GetType() == types.PlanTypePublic {
				startTime := ctx.BlockTime()
				endTime := startTime.AddDate(0, simtypes.RandIntBetween(r, 1, 28), 0)

				switch plan := p.(type) {
				case *types.FixedAmountPlan:
					req.PlanId = plan.GetId()
					req.Name = "simulation-test-" + simtypes.RandStringOfLength(r, 5)
					req.FarmingPoolAddress = plan.GetFarmingPoolAddress().String()
					req.TerminationAddress = plan.GetTerminationAddress().String()
					req.StakingCoinWeights = plan.GetStakingCoinWeights()
					req.StartTime = &startTime
					req.EndTime = &endTime
					req.EpochAmount = sdk.NewCoins(sdk.NewInt64Coin(poolCoins[r.Intn(3)].Denom, int64(simtypes.RandIntBetween(r, 10_000_000, 1_000_000_000))))
				case *types.RatioPlan:
					req.PlanId = plan.GetId()
					req.Name = "simulation-test-" + simtypes.RandStringOfLength(r, 5)
					req.FarmingPoolAddress = plan.GetFarmingPoolAddress().String()
					req.TerminationAddress = plan.GetTerminationAddress().String()
					req.StakingCoinWeights = plan.GetStakingCoinWeights()
					req.StartTime = &startTime
					req.EndTime = &endTime
					req.EpochRatio = sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 5)), 1)
				}
				break
			}
		}

		if req.PlanId == 0 {
			return nil
		}

		updateRequests := []*types.UpdateRequestProposal{req}

		return types.NewPublicPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			[]*types.AddRequestProposal{},
			updateRequests,
			[]*types.DeleteRequestProposal{},
		)
	}
}

// SimulateDeletePublicPlanProposal generates random public plan proposal content
func SimulateDeletePublicPlanProposal(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		params := k.GetParams(ctx)
		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return nil
		}

		req := &types.DeleteRequestProposal{}

		plans := k.GetPlans(ctx)
		for _, p := range plans {
			if p.GetType() == types.PlanTypePublic {
				req.PlanId = p.GetId()
				break
			}
		}

		if req.PlanId == 0 {
			return nil
		}

		deleteRequest := []*types.DeleteRequestProposal{req}

		return types.NewPublicPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			[]*types.AddRequestProposal{},
			[]*types.UpdateRequestProposal{},
			deleteRequest,
		)
	}
}

// genAddRequestProposals returns randomized add request proposals.
func genAddRequestProposals(r *rand.Rand, ctx sdk.Context, simAccount simtypes.Account, poolCoins sdk.Coins) []*types.AddRequestProposal {
	ranProposals := make([]*types.AddRequestProposal, 0)

	// generate random number of proposals with random values of each parameter
	// it generates a fixed amount plan if pseudo-random integer is an even number then and
	// it generates a ratio plan if it is an odo number
	for i := 0; i < simtypes.RandIntBetween(r, 1, 3); i++ {
		req := &types.AddRequestProposal{}
		if r.Int()%2 == 0 {
			req.Name = "simulation-test-" + simtypes.RandStringOfLength(r, 5)
			req.FarmingPoolAddress = simAccount.Address.String()
			req.TerminationAddress = simAccount.Address.String()
			req.StakingCoinWeights = sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1))
			req.StartTime = ctx.BlockTime()
			req.EndTime = ctx.BlockTime().AddDate(0, simtypes.RandIntBetween(r, 1, 28), 0)
			req.EpochAmount = sdk.NewCoins(sdk.NewInt64Coin(poolCoins[r.Intn(3)].Denom, int64(simtypes.RandIntBetween(r, 10_000_000, 1_000_000_000))))
		} else {
			req.Name = "simulation-test-" + simtypes.RandStringOfLength(r, 5)
			req.FarmingPoolAddress = simAccount.Address.String()
			req.TerminationAddress = simAccount.Address.String()
			req.StakingCoinWeights = sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1))
			req.StartTime = ctx.BlockTime()
			req.EndTime = ctx.BlockTime().AddDate(0, simtypes.RandIntBetween(r, 1, 28), 0)
			req.EpochRatio = sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 5)), 1)
		}

		ranProposals = append(ranProposals, req)
	}
	return ranProposals
}
