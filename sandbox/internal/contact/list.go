package contact

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// What the backoffice reads and changes of the contact requests: one page of
// them as a filter keeps them, one whole, and — for a root — archiving,
// unarchiving and deleting them.

// List is one page of the contact requests filter keeps, newest first.
func List(sandbox *api.Sandbox, filter Filter, page int, limit int) ([]contact_db.ContactRequestRecord, visualization.Paging, error) {
	kept, err := Selected(sandbox, filter)
	if err != nil {
		return nil, visualization.Paging{}, err
	}
	paging, from, to := visualization.Paginate(sandbox, len(kept), page, limit)
	return kept[from:to], paging, nil
}

// Selected is every contact request filter keeps, newest first.
func Selected(sandbox *api.Sandbox, filter Filter) ([]contact_db.ContactRequestRecord, error) {
	all, err := contact_db.ListContactRequestsBetween(sandbox, contact_db.New(sandbox), filter.AtMin, filter.AtMax)
	if err != nil {
		return nil, err
	}
	needle := sandbox.Deps.StringsDeps.ToLower(filter.Contains)
	kept := []contact_db.ContactRequestRecord{}
	for _, request := range all {
		if keeps(sandbox, filter, needle, request.Archived, request.Country, request.Body) {
			kept = append(kept, request)
		}
	}
	sandbox.Deps.SortDeps.SliceStable(kept, func(i int, j int) bool {
		if kept[i].At != kept[j].At {
			return kept[i].At > kept[j].At
		}
		return kept[i].Id > kept[j].Id
	})
	return kept, nil
}

// Find is the contact request whose id is ref, false when there is none.
func Find(sandbox *api.Sandbox, ref string) (contact_db.ContactRequestRecord, bool) {
	ref = sandbox.Deps.StringsDeps.ToLower(sandbox.Deps.StringsDeps.TrimSpace(ref))
	if ref == "" {
		return contact_db.ContactRequestRecord{}, false
	}
	return contact_db.New(sandbox).FindContactRequestByRef(ref)
}

// findAll is the contact requests whose ids are refs — each once, the ids no
// contact request holds left out — or the refusal of a list that names none,
// or more than MaxIds: 404 when none of them is found.
func findAll(sandbox *api.Sandbox, refs []string) ([]contact_db.ContactRequestRecord, *Refusal) {
	seen := map[string]bool{}
	found := []contact_db.ContactRequestRecord{}
	named := 0
	for _, ref := range refs {
		ref = sandbox.Deps.StringsDeps.ToLower(sandbox.Deps.StringsDeps.TrimSpace(ref))
		if ref == "" || seen[ref] {
			continue
		}
		seen[ref] = true
		named++
		if named > MaxIds {
			return nil, refuse("ids", sandbox.Deps.StdDeps.Sprintf("Name at most %d contact requests at once.", MaxIds))
		}
		if request, ok := Find(sandbox, ref); ok {
			found = append(found, request)
		}
	}
	if named == 0 {
		return nil, refuse("ids", "Select at least one contact request.")
	}
	if len(found) == 0 {
		return nil, &Refusal{Status: api.StatusNotFound, Field: "ids", Message: "No contact request has these ids."}
	}
	return found, nil
}

// SetArchived archives — or, with archived false, brings back to the inbox —
// the contact requests whose ids are refs, and answers how many changed.
func SetArchived(sandbox *api.Sandbox, refs []string, archived bool) (int, *Refusal, error) {
	found, refusal := findAll(sandbox, refs)
	if refusal != nil {
		return 0, refusal, nil
	}
	value := int64(0)
	if archived {
		value = 1
	}
	db := contact_db.New(sandbox)
	changed := 0
	for _, request := range found {
		if request.Archived == value {
			continue
		}
		if err := db.SetContactRequestArchived(request.Id, value); err != nil {
			return changed, nil, err
		}
		changed++
	}
	return changed, nil, nil
}

// Remove removes the contact requests whose ids are refs, or every contact
// request before the day before — exactly one of the two — and answers how
// many it removed.
func Remove(sandbox *api.Sandbox, refs []string, before string) (int, *Refusal, error) {
	before = sandbox.Deps.StringsDeps.TrimSpace(before)
	named := false
	for _, ref := range refs {
		if sandbox.Deps.StringsDeps.TrimSpace(ref) != "" {
			named = true
		}
	}
	if named == (before != "") {
		return 0, refuse("ids", "Send the ids of the contact requests, or a date to remove every contact request before it."), nil
	}
	if before != "" {
		cutoff, ok := visualization.ParseDay(sandbox, before)
		if !ok {
			return 0, refuse("before", "before is a date spelled YYYY-MM-DD, such as 2026-10-14."), nil
		}
		removed, err := contact_db.RemoveContactRequestsBefore(sandbox, contact_db.New(sandbox), cutoff)
		return removed, nil, err
	}
	found, refusal := findAll(sandbox, refs)
	if refusal != nil {
		return 0, refusal, nil
	}
	db := contact_db.New(sandbox)
	for removed, request := range found {
		if err := db.RemoveContactRequest(request.Id); err != nil {
			return removed, nil, err
		}
	}
	return len(found), nil, nil
}

// HomeCounts is how many contact requests are in the inbox, and how many
// arrived today, for the home page of the backoffice.
func HomeCounts(sandbox *api.Sandbox) (int, int, error) {
	todayStart, _ := visualization.ParseDay(sandbox, visualization.Today(sandbox))
	return contact_db.CountContactRequests(sandbox, contact_db.New(sandbox), todayStart)
}

// BackTo is where a form of the contact pages answers 303 to, with notice and
// count in its query string: back when it is the list, one contact request or
// the settings — a path of the backoffice, never another site — the list
// otherwise. With gone true, the contact requests were deleted, so a back to
// one of them goes to the list instead.
func BackTo(sandbox *api.Sandbox, back string, gone bool, notice string, count int) string {
	strings := sandbox.Deps.StringsDeps
	under := func(path string) bool {
		return back == path || strings.HasPrefix(back, path+"?")
	}
	safe := !strings.ContainsAny(back, "\r\n\\#") && (under(RequestsPath) || under(RequestPath) || under(SettingsPath))
	if !safe || (gone && under(RequestPath)) {
		back = RequestsPath
	}
	separator := "?"
	if strings.Contains(back, "?") {
		separator = "&"
	}
	return back + separator + "notice=" + notice + "&count=" + strings.FormatInt(int64(count), 10)
}

// MaxExport is how many contact requests one export holds at most.
const MaxExport = 10000

// Export is the contact requests an access or a portability request asks
// for: those of refs, or — with no refs — every one filter keeps, newest
// first. More than MaxExport is refused.
func Export(sandbox *api.Sandbox, filter Filter, refs []string) ([]contact_db.ContactRequestRecord, *Refusal, error) {
	named := []string{}
	for _, ref := range refs {
		if sandbox.Deps.StringsDeps.TrimSpace(ref) != "" {
			named = append(named, ref)
		}
	}
	if len(named) > 0 {
		if len(named) > MaxExport {
			return nil, refuse("ids", "Export at most 10000 contact requests at once."), nil
		}
		found := []contact_db.ContactRequestRecord{}
		seen := map[string]bool{}
		for _, ref := range named {
			if request, ok := Find(sandbox, ref); ok && !seen[request.Ref] {
				seen[request.Ref] = true
				found = append(found, request)
			}
		}
		return found, nil, nil
	}
	kept, err := Selected(sandbox, filter)
	if err != nil {
		return nil, nil, err
	}
	if len(kept) > MaxExport {
		return nil, refuse("", "The filter keeps more than 10000 contact requests: narrow it."), nil
	}
	return kept, nil, nil
}
