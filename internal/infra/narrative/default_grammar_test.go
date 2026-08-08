package narrative_test

import (
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/narrative"
	infranarrative "github.com/thalesraymond/world-generation-go/internal/infra/narrative"
)

// topLevelCategories are the event-category rules that must satisfy the
// viability invariant: at least one alternative whose direct variables are
// a subset of the guaranteed context passed by the chronicle caller. Any
// other rule in the grammar is treated as a sub-rule (referenced only via
// NonTerminal from a top-level rule) and is not required to declare a
// guaranteed context.
var topLevelCategories = []string{
	"Birth", "Death", "Marriage", "RoleTransition", "Settlement",
	"Conflict", "Politics", "Discovery",
	"Expansion", "Raid", "Conquest", "Diplomacy", "Economy", "AgentAction",
}

// figureRuleNames are the .figure variants of the figure-driven categories.
// These rules are exempt from the viability check because the caller only
// attempts them when the figure is known to be present, and the chain
// falls back to the base rule on ErrNoEligibleAlternative / ErrRuleNotFound.
var figureRuleNames = []string{"Conflict.figure", "Politics.figure", "Discovery.figure"}

// deadRuleNames are rules with no producing timeline category. The chronicle
// caller falls back to event.Description on ErrRuleNotFound, so removing
// these rules is safe.
var deadRuleNames = []string{"Disaster", "Succession", "ReputationChange"}

func guaranteedContextFor(ruleName string) map[string]struct{} {
	switch ruleName {
	case "Birth", "Death", "Marriage", "RoleTransition", "Settlement":
		return set("year", "description", "SettlementName", "FigureName")
	case "Conflict", "Politics", "Discovery":
		return set("year", "description")
	case "Expansion", "Raid", "Conquest", "Diplomacy", "Economy", "AgentAction":
		return set("year", "description", "SettlementName", "ActionType", "Outcome", "TargetSettlement")
	}
	return nil
}

func set(items ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, it := range items {
		m[it] = struct{}{}
	}
	return m
}

// subset reports whether every element of sub is in sup.
func subset(sub, sup map[string]struct{}) bool {
	for k := range sub {
		if _, ok := sup[k]; !ok {
			return false
		}
	}
	return true
}

// isFigureRule reports whether a rule name is a figure-driven variant
// (e.g. "Conflict.figure"). The base rules of these names are figure-free.
func isFigureRule(name string) bool {
	return strings.HasSuffix(name, ".figure")
}

// hasYearReference reports whether the alternative references $year anywhere
// (direct variable only).
func hasYearReference(alt narrative.Alternative) bool {
	for _, n := range narrative.AlternativeVariables(alt) {
		if n == "year" {
			return true
		}
	}
	return false
}

// hasRoleReference reports whether the alternative references $FigureRole
// (direct variable only).
func hasRoleReference(alt narrative.Alternative) bool {
	for _, n := range narrative.AlternativeVariables(alt) {
		if n == "FigureRole" {
			return true
		}
	}
	return false
}

// descriptionPlacement returns the indices of $description symbols in the
// alternative that are placed as mid-sentence noun phrases. The allowed
// positions are:
//   - first symbol of the alt (a standalone sentence)
//   - immediately preceded by a terminal whose last non-whitespace char is
//     one of , . : — (a clause-boundary introducer)
//
// A $description embedded between words (e.g. "The people of " $description
// " rejoiced") fails the check because the preceding terminal is "of " — a
// noun-phrase preposition, not a clause boundary.
func descriptionPlacement(alt narrative.Alternative) []int {
	var badPositions []int
	for i, sym := range alt {
		v, ok := sym.(narrative.Variable)
		if !ok || v.Name != "description" {
			continue
		}
		if i == 0 {
			continue // first symbol: standalone sentence body
		}
		prev, ok := alt[i-1].(narrative.Terminal)
		if !ok {
			badPositions = append(badPositions, i)
			continue
		}
		if !isClauseBoundaryTerminal(prev.Text) {
			badPositions = append(badPositions, i)
		}
	}
	return badPositions
}

// isClauseBoundaryTerminal reports whether the terminal text ends (after
// trimming trailing whitespace) with one of the clause-boundary punctuation
// characters: , . : —.
func isClauseBoundaryTerminal(text string) bool {
	trimmed := strings.TrimRight(text, " \t")
	if trimmed == "" {
		return false
	}
	runes := []rune(trimmed)
	last := runes[len(runes)-1]
	return last == ',' || last == '.' || last == ':' || last == '—'
}

func TestDefaultGrammar_StaticInvariants(t *testing.T) {
	g, err := narrative.Parse(infranarrative.DefaultGrammar)
	if err != nil {
		t.Fatalf("parse default grammar: %v", err)
	}

	// 1. Every top-level category has at least one alternative whose direct
	//    variables are a subset of the guaranteed context.
	for _, name := range topLevelCategories {
		rule, ok := g.Rules[name]
		if !ok {
			t.Errorf("top-level rule %q missing from DefaultGrammar", name)
			continue
		}
		gctx := guaranteedContextFor(name)
		if gctx == nil {
			t.Errorf("top-level rule %q has no declared guaranteed context", name)
			continue
		}
		var eligible int
		for _, alt := range rule.Alternatives {
			vars := narrative.AlternativeVariables(alt)
			if subset(set(vars...), gctx) {
				eligible++
			}
		}
		if eligible == 0 {
			t.Errorf("rule %q: no alternative is eligible under guaranteed context %v", name, keys(gctx))
		}
	}

	// 2. RoleTransition has at least one role-free alternative.
	rt, ok := g.Rules["RoleTransition"]
	if !ok {
		t.Fatal("RoleTransition rule missing from DefaultGrammar")
	} else {
		var hasRoleFree bool
		for _, alt := range rt.Alternatives {
			if !hasRoleReference(alt) {
				hasRoleFree = true
				break
			}
		}
		if !hasRoleFree {
			t.Error("RoleTransition has no role-free alternative; empty FigureRole will make the whole rule ineligible")
		}
	}

	// 3. Every .figure alternative references $year (chronology invariant).
	for _, name := range figureRuleNames {
		rule, ok := g.Rules[name]
		if !ok {
			t.Errorf("figure rule %q missing from DefaultGrammar", name)
			continue
		}
		for i, alt := range rule.Alternatives {
			if !hasYearReference(alt) {
				t.Errorf("%s alternative %d does not reference $year (chronology defect)", name, i)
			}
		}
	}

	// 4. No alternative embeds $description as a mid-sentence noun phrase.
	for name, rule := range g.Rules {
		for i, alt := range rule.Alternatives {
			if bad := descriptionPlacement(alt); len(bad) > 0 {
				t.Errorf("%s alternative %d embeds $description as a mid-sentence noun phrase at positions %v", name, i, bad)
			}
		}
	}

	// 5. Dead rules (no producing timeline category) are absent.
	for _, dead := range deadRuleNames {
		if _, ok := g.Rules[dead]; ok {
			t.Errorf("dead rule %q is present in DefaultGrammar; it should be removed (no producing category)", dead)
		}
	}
}

func TestDefaultGrammar_AgentCategoriesEliminateSubjectEcho(t *testing.T) {
	g, err := narrative.Parse(infranarrative.DefaultGrammar)
	if err != nil {
		t.Fatalf("parse default grammar: %v", err)
	}
	// The agent categories (Expansion, Raid, Conquest, Diplomacy, Economy)
	// receive a complete-sentence $Outcome that already includes the
	// subject (the source settlement). Templates that re-state the subject
	// via $SettlementName inside the same sentence produce the
	// outcome-echo defect (e.g. "In 7, the people of Alpha raised new
	// banners beyond their walls: Alpha founded Newhold."). Each
	// alternative that references $SettlementName is therefore forbidden.
	for _, name := range []string{"Expansion", "Raid", "Conquest", "Diplomacy", "Economy"} {
		rule, ok := g.Rules[name]
		if !ok {
			continue
		}
		for i, alt := range rule.Alternatives {
			for _, v := range narrative.AlternativeVariables(alt) {
				if v == "SettlementName" {
					t.Errorf("%s alternative %d references $SettlementName, which duplicates the subject already in $Outcome (outcome-echo defect)", name, i)
				}
			}
		}
	}
}

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
